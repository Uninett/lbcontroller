package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	//"github.com/koki/json"
	"k8s.io/apimachinery/pkg/util/json"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
		return response, nil
	}

	//find the status of the load balancer
	fmt.Printf("TODO: find status of laod balancer for service %s\n", request.Service.Name)

	fmt.Println("TODO: add load balancers")

	fmt.Println("TODO: get annotations from the loadbalancers API")

	fmt.Println("TODO: generate NetworkPolicy")

	// Generate desired NetworkPolicy
	netpol := netv1.NetworkPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.k8s.io/v1",
			Kind:       "NetworkPolicy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: request.Service.Name + "-lb",
		},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			Ingress: []netv1.NetworkPolicyIngressRule{
				{
					Ports: []netv1.NetworkPolicyPort{
						{
							//Protocol: &v1.Protocol("TCP"),
						},
					},
					From: []netv1.NetworkPolicyPeer{},
				},
			},
			//Egress:      {},
			//PolicyTypes: {},
		},
	}
	response.Labels["this"] = "that"
	response.Annotations["these"] = "those"
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

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/sync", syncHandler).Methods("POST")
	loggedRouter := handlers.LoggingHandler(os.Stdout, router)
	log.Fatal(http.ListenAndServe(":8080", loggedRouter))
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
