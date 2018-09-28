package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	//"github.com/koki/json"
	"k8s.io/apimachinery/pkg/util/intstr"
	//"k8s.io/apimachinery/pkg/util/json"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const defaultCluster = "nird"

var (
	lbpeersString = kingpin.Flag("peers", "The load babalancers IPs, comma separated in CIDR form").Required().Envar("LBC_PEERS").String()
	lbendpoint    = kingpin.Flag("endpoint", "The load balancer controller API endpoint").Required().Envar("LBC_ENDPOINT").String()
	cluster       = kingpin.Flag("clustername", "The name of the Kubernetes cluster").Default("nird").Envar("LBC_CLUSTER_NAME").String()
	token         = kingpin.Flag("token", "Authentication token to access the load balancer API").Required().Envar("LBC_TOKEN").String()
	lbpeers       []string // split strings of lbpeersString
)

func init() {
	//	fmt.Printf("%#v\n", os.Environ())
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	kingpin.Parse()

	lbpeers = strings.Split(*lbpeersString, ",")

	router := mux.NewRouter()
	router.HandleFunc("/sync", syncHandler).Methods("POST")
	loggedRouter := handlers.LoggingHandler(os.Stdout, router)
	log.Fatal(http.ListenAndServe(":8080", loggedRouter))
}

//SyncRequest is the request from the metacontroller
type SyncRequest struct {
	Controller  json.RawMessage                           `json:"controller"`
	Service     v1.Service                                `json:"object"`
	Attachments map[string]map[string]netv1.NetworkPolicy `json:"attachments"`
}

//SyncResponse is the response to the metacontroller
type SyncResponse struct {
	Labels      map[string]string     `json:"labels"`
	Annotations map[string]string     `json:"annotations"`
	Attachments []netv1.NetworkPolicy `json:"attachments"`
}

func sync(request *SyncRequest) (*SyncResponse, error) {
	response := &SyncResponse{}
	response.Labels = make(map[string]string)
	response.Annotations = make(map[string]string)
	response.Attachments = make([]netv1.NetworkPolicy, 0, 1)

	if request.Service.Spec.Type != v1.ServiceTypeLoadBalancer {
		log.Println("not a loadbalancer service")
		//TODO (gta) empty response? keep labels and annotations!
		return response, nil
	}

	//get protocol and ports from k8s service
	//TODO(gta)log that we cannot support mutliple protocols services.
	//And do nothing as earlier for non loadbalancer services.
	svcPorts, svcProto, err := getPortsProto(request.Service)
	protoString := strings.ToLower(string(svcProto))
	if err != nil {
		log.Println("ERROR: cannot specify both TCP and UDP protocols in the same K8s Service. Create two services.")
		return response, nil
	}

	var (
		namespace   = request.Service.Namespace
		serviceName = request.Service.Name
	)
	//serviceLbKey := strings.Join([]string{*cluster, namespace, serviceName, protoString}, "-")
	serviceLbKey := strings.Join([]string{*cluster, namespace, serviceName, protoString}, "")

	log.Println("sync load balancer service")

	lbService := newlbcontrollerService(request.Service, serviceLbKey, protoString)

	ingress, err := SyncService(lbService, *lbendpoint, *token)
	if err != nil {
		return response, errors.Wrap(err, "Could not create load balancer service")
	}

	log.Printf("Created/updated load balancer with ingress: %v\n", ingress)

	//TODO this annotation might be not optimal
	for _, in := range ingress {
		response.Annotations[in.Hostname] = in.IP
	}

	log.Println("generate NetworkPolicy")

	netpol := newNetworkPolicy(request.Service, ingress, svcProto, svcPorts)

	response.Labels["LoadBalncer"] = "true" //TODO change this in something more useful?

	response.Attachments = append(response.Attachments, netpol)

	return response, nil
}

func syncHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	request := &SyncRequest{}
	if err := json.Unmarshal(body, request); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	response, err := sync(request)
	if err != nil {
		log.Printf("%v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	body, err = json.Marshal(&response)
	//log.Println(string(body))
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func syncLoadBalancerService(v1.Service, Service) error {
	log.Printf("TODO syncLoadBalancerService")
	return nil
}

//TODO(gta) if there are both tcp and udp specified log an error and do nothing.
func getPortsProto(service v1.Service) ([]int32, v1.Protocol, error) {
	var (
		svcProto           = v1.ProtocolTCP //default
		svcPorts           = []int32{}
		tcpProto, udpProto bool
	)

	for _, p := range service.Spec.Ports {
		if p.Protocol == v1.ProtocolUDP {
			udpProto = true
			svcProto = v1.ProtocolUDP
		}
		if p.Protocol == v1.ProtocolTCP {
			tcpProto = true
		}
		svcPorts = append(svcPorts, p.Port)
	}
	if tcpProto && udpProto { //too many protocols
		return nil, "", errors.New("both TCP and UDP specified in service")
	}
	return svcPorts, svcProto, nil
}

func newlbcontrollerService(ks v1.Service, key, protocol string) Service {
	svc := Service{}
	svc.Type = ServiceType(protocol)
	svc.Metadata.Name = key
	cfg := Config{
		Method:           "least_conn",
		UpstreamMaxConns: 100,
	}
	cfg.Backends = backends //TODO this is hardcoded for now
	if len(ks.Spec.LoadBalancerSourceRanges) != 0 {
		cfg.ACL = ks.Spec.LoadBalancerSourceRanges
	}
	if ks.Spec.HealthCheckNodePort != 0 {
		cfg.HealthCheck.Port = ks.Spec.HealthCheckNodePort
	} else if len(ks.Spec.Ports) > 0 {
		cfg.HealthCheck.Port = ks.Spec.Ports[0].NodePort
	}

	cfg.Ports = make(map[string]int32)
	for _, p := range ks.Spec.Ports {
		if string(p.Protocol) == protocol {
			port := fmt.Sprint(p.Port)
			cfg.Ports[port] = int32(p.NodePort)
		}
	}

	return svc
}

func newNetworkPolicy(ksvc v1.Service, ingress []v1.LoadBalancerIngress, proto v1.Protocol, ports []int32) netv1.NetworkPolicy {

	netPolPorts := []netv1.NetworkPolicyPort{}
	for _, p := range ports {
		port := netv1.NetworkPolicyPort{
			Protocol: &proto,
			Port: &intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: p,
			},
		}
		netPolPorts = append(netPolPorts, port)
	}

	netPolPeers := []netv1.NetworkPolicyPeer{}
	for _, peer := range lbpeers {
		netin := netv1.NetworkPolicyPeer{
			IPBlock: &netv1.IPBlock{
				CIDR: peer,
			},
		}
		netPolPeers = append(netPolPeers, netin)
	}

	netpol := netv1.NetworkPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.k8s.io/v1",
			Kind:       "NetworkPolicy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: ksvc.Name + "-lb",
		},
		Spec: netv1.NetworkPolicySpec{
			PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeIngress},
			PodSelector: metav1.LabelSelector{
				MatchLabels: ksvc.Spec.Selector,
			},
			Ingress: []netv1.NetworkPolicyIngressRule{
				{
					Ports: netPolPorts,
					From:  netPolPeers,
				},
			},
		},
	}
	return netpol
}

//TODO put this in a config file
var backends = []Backend{
	{
		Host:  "tos-spw01.nird.sigma2.no",
		Addrs: []string{"193.156.11.24", "2001:700:4a00:11::1024"},
	},
	{
		Host:  "tos-spw02.nird.sigma2.no",
		Addrs: []string{"193.156.11.25", "2001:700:4a00:11::1025"},
	},
	{
		Host:  "tos-spw03.nird.sigma2.no",
		Addrs: []string{"193.156.11.26", "2001:700:4a00:11::1026"},
	},
	{
		Host:  "tos-spw04.nird.sigma2.no",
		Addrs: []string{"193.156.11.27", "2001:700:4a00:11::1027"},
	},
	{
		Host:  "tos-spw05.nird.sigma2.no",
		Addrs: []string{"193.156.11.28", "2001:700:4a00:11::1028"},
	},
	{
		Host:  "tos-spw06.nird.sigma2.no",
		Addrs: []string{"193.156.11.29", "2001:700:4a00:11::1029"},
	},
	{
		Host:  "tos-spw07.nird.sigma2.no",
		Addrs: []string{"193.156.11.30", "2001:700:4a00:11::1030"},
	},
}
