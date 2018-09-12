// +build ignore
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/koki/json"

	"github.com/UNINETT/lbcontroller"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

var (
	services = map[string]lbcontroller.Service{}
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/services", listServices).Methods("GET")
	router.HandleFunc("/services/{name}", getService).Methods("GET")
	router.HandleFunc("/services/{name}", syncService).Methods("PUT")
	router.HandleFunc("/services/{name}", deleteService).Methods("DELETE")

	router.HandleFunc("/ingress", getIngress).Methods("GET")

	loggedRouter := handlers.LoggingHandler(os.Stdout, router)
	http.ListenAndServe(":8080", loggedRouter)
}

func getService(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(req)
	name := vars["name"]

	if len(name) == 0 {
		log.Println("need to specify a resource name")
		http.Error(res, "need to specify a resource name", http.StatusInternalServerError)
		return
	}

	outData, present := services[name]
	if !present {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	outgoingJSON, error := json.Marshal(outData)
	if error != nil {
		log.Println(error.Error())
		http.Error(res, error.Error(), http.StatusInternalServerError)
		return
	}
	location := "http://" + req.Host + "/ingress"
	res.Header().Add("Location", location)
	fmt.Fprint(res, string(outgoingJSON))

}

func listServices(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	for _, svc := range services {

		outgoingJSON, error := json.Marshal(svc)
		if error != nil {
			log.Println(error.Error())
			http.Error(res, error.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(res, string(outgoingJSON))
	}
}
func syncService(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(req)
	name := vars["name"]

	newSvc := lbcontroller.Service{}
	decoder := json.NewDecoder(req.Body)
	error := decoder.Decode(&newSvc)
	if error != nil {
		log.Println(error.Error())
		http.Error(res, error.Error(), http.StatusInternalServerError)
		return
	}
	if name != newSvc.Metadata.Name {
		err := errors.Errorf("Name of service inconsistent expected %s got %s\n", name, newSvc.Metadata.Name)
		log.Printf("%v\n", err)
		http.Error(res, error.Error(), http.StatusInternalServerError)
		return
	}
	now := time.Now()
	_, present := services[name]
	if present {
		//err := errors.Errorf("Service %s already present, updating\n", newSvc.Metadata.Name)
		//log.Println(err)
		log.Printf("Service %s already present, updating\n", newSvc.Metadata.Name)
		newSvc.Metadata.UpdatedAt = now
		//http.Error(res, err.Error(), http.StatusInternalServerError)
		//return
	} else {
		newSvc.Metadata.CreatedAt = now
		newSvc.Metadata.UpdatedAt = now
	}
	services[name] = newSvc
	location := "http://" + req.Host + "/ingress"
	res.Header().Add("Location", location)
	//res.WriteHeader(http.StatusNoContent)
	if !present {
		res.WriteHeader(http.StatusCreated)
	}

}

func deleteService(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(req)
	name := vars["name"]

	delete(services, name)
	res.WriteHeader(http.StatusNoContent)

}

func getIngress(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")

	ret := `[
		{
		 "ip": "127.0.0.1",
		 "hostname": "ingress.host.com"
		}   
	 ]`

	fmt.Fprint(res, ret)
}
