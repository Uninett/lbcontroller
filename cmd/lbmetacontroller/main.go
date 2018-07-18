package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	//"github.com/koki/json"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/json"

	"github.com/UNINETT/lbcontroller"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const defaultCluster = "nird"

var (
	lbendpoint string
	cluster    string
)

func main() {
	lbendpoint = os.Getenv("lbcontroller_ENDPOINT")
	if lbendpoint == "" {
		panic("no load balancer endpoint defined")
	}
	cluster = os.Getenv("lbcontroller_CLUSTER")
	if cluster == "" {
		log.Printf("No clustrer name defined, defalt name is given: %s", defaultCluster)
		cluster = defaultCluster
	}
	router := mux.NewRouter()
	router.HandleFunc("/sync", syncHandler).Methods("POST")
	loggedRouter := handlers.LoggingHandler(os.Stdout, router)
	log.Fatal(http.ListenAndServe(":8080", loggedRouter))
}

//SyncRequest is the request from the metacontroller
type SyncRequest struct {
	Controller  RawMessage                                `json:"controller"`
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
		fmt.Println("not a loadbalancer service")
		//TODO (gta) empty response?
		return response, nil
	}

	//TCP or UDP?
	var (
		svcProto = v1.ProtocolTCP //default
		svcPorts = []int32{}
	)

	for _, p := range request.Service.Spec.Ports {
		if p.Protocol == v1.ProtocolUDP {
			svcProto = v1.ProtocolUDP
		}
		svcPorts = append(svcPorts, p.Port)
	}

	var (
		namespace   = request.Service.Namespace
		serviceName = request.Service.Name
	)
	serviceLbKey := strings.Join([]string{cluster, namespace, serviceName, string(svcProto)}, "-")

	//find the status of the load balancer
	log.Printf("finding status of laod balancer for service %s\n", request.Service.Name)
	_, found, err := lbcontroller.GetService(serviceLbKey, lbendpoint)
	if err != nil {
		return response, errors.Wrapf(err, "ERROR deleting service with key %s\n", serviceLbKey)
	}
	if found {
		log.Printf("service %s present", serviceLbKey)
		//Check if the two versions are different, if not do nothing.
		//If they are different maybe betteto delete and recreate LB service.
		//TODO synch the load balancer and the service
		log.Println("TODO: synch the load balancer and the service")
	}

	log.Println("add load balancer")

	lbService := newlbcontrollerService(request.Service, serviceLbKey, string(svcProto))

	ingress, err := lbcontroller.NewService(lbService, lbendpoint)
	if err != nil {
		return response, errors.Wrap(err, "Could not create load balancer service")
	}

	log.Printf("Created loab balancer with ingress: %v\n", ingress)

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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	request := &SyncRequest{}
	if err := json.Unmarshal(body, request); err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	response, err := sync(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	body, err = json.Marshal(&response)
	fmt.Println(string(body))
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

// RawMessage is a raw encoded JSON value.
// It implements Marshaler and Unmarshaler and can
// be used to delay JSON decoding or precompute a JSON encoding.
type RawMessage []byte

// MarshalJSON returns m as the JSON encoding of m.
func (m RawMessage) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	return m, nil
}

// UnmarshalJSON sets *m to a copy of data.
func (m *RawMessage) UnmarshalJSON(data []byte) error {
	if m == nil {
		return errors.New("json.RawMessage: UnmarshalJSON on nil pointer")
	}
	*m = append((*m)[0:0], data...)
	return nil
}

func syncLoadBalancerService(v1.Service, lbcontroller.Service) error {
	log.Printf("TODO syncLoadBalancerService")
	return nil
}

func getPortsProto(service v1.Service) ([]int32, v1.Protocol) {
	var (
		svcProto = v1.ProtocolTCP //default
		svcPorts = []int32{}
	)

	for _, p := range service.Spec.Ports {
		if p.Protocol == v1.ProtocolUDP {
			svcProto = v1.ProtocolUDP
		}
		svcPorts = append(svcPorts, p.Port)
	}
	return svcPorts, svcProto
}

func newlbcontrollerService(ks v1.Service, key, protocol string) lbcontroller.Service {
	svc := lbcontroller.Service{}
	svc.Type = lbcontroller.ServiceType(strings.ToLower(protocol))
	svc.Metadata.Name = key
	cfg := lbcontroller.Config{
		Method:           "least_conn",
		UpstreamMaxConns: 100,
	}
	cfg.Backends = backends
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
	for _, in := range ingress {
		netin := netv1.NetworkPolicyPeer{
			IPBlock: &netv1.IPBlock{
				CIDR: in.IP, // TODO is this OK
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
				//TODO same as the service?
				MatchLabels: ksvc.Spec.Selector,
			},
			Ingress: []netv1.NetworkPolicyIngressRule{
				{
					Ports: netPolPorts,
					From:  netPolPeers,
				},
			},
			//Egress:      {},
		},
	}
	return netpol
}

//TODO put this in a config file
var backends = []lbcontroller.Backend{
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
