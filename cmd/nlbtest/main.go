package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"

	"github.com/koki/json"

	"github.com/folago/nlb"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var (
	services = map[string]nlb.Service{}
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/services", listServices).Methods("GET")
	router.HandleFunc("/services/{name}", getService).Methods("GET")
	router.HandleFunc("/services", newService).Methods("POST")
	router.HandleFunc("/services/{name}", editService).Methods("DELETE", "PATCH")

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

func newService(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")

	newSvc := nlb.Service{}
	decoder := json.NewDecoder(req.Body)
	error := decoder.Decode(&newSvc)
	if error != nil {
		log.Println(error.Error())
		http.Error(res, error.Error(), http.StatusInternalServerError)
		return
	}
	_, present := services[newSvc.Metadata.Name]
	if present {
		err := errors.Errorf("Service %s already present\n", newSvc.Metadata.Name)
		log.Println(err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	now := time.Now()
	newSvc.Metadata.CreatedAt = now
	newSvc.Metadata.UpdatedAt = now
	services[newSvc.Metadata.Name] = newSvc
	location := "http://" + req.Host + "/ingress"
	res.Header().Add("Location", location)
	//res.WriteHeader(http.StatusNoContent)
	res.WriteHeader(http.StatusCreated)

}

func editService(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(req)
	name := vars["name"]

	switch req.Method {
	case "DELETE":
		delete(services, name)
		res.WriteHeader(http.StatusNoContent)
	case "PATCH":
		editingsvc, ok := services[name]
		if !ok {
			res.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(res, "Service %s not found", name)
			return
		}
		svc := nlb.Service{}
		decoder := json.NewDecoder(req.Body)
		error := decoder.Decode(&svc)
		if error != nil {
			log.Println(error.Error())
			http.Error(res, error.Error(), http.StatusInternalServerError)
			return
		}
		svc.Metadata.CreatedAt = editingsvc.Metadata.CreatedAt
		svc.Metadata.UpdatedAt = time.Now()
		services[name] = svc

		res.WriteHeader(http.StatusNoContent)
	}
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
