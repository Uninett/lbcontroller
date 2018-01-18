package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/koki/json"

	"github.com/folago/nlb"
	"github.com/gorilla/mux"
)

var (
	frontends = map[string]nlb.Frontend{}
	services  = map[string]nlb.Service{}
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/frontends", listFrontends).Methods("GET")
	router.HandleFunc("/frontends/{name}", listFrontends).Methods("GET")
	router.HandleFunc("/frontends", newFrontend).Methods("POST")
	router.HandleFunc("/frontends/{name}", editFrontends).Methods("PUT", "DELETE", "PATCH")

	router.HandleFunc("/services", listServices).Methods("GET")
	router.HandleFunc("/services/{name}", listServices).Methods("GET")
	router.HandleFunc("/services", newService).Methods("POST")
	router.HandleFunc("/services/{name}", editService).Methods("PUT", "DELETE", "PATCH")

	http.ListenAndServe(":8080", router)
}

func listFrontends(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(req)
	name := vars["name"]

	var outData interface{}
	if len(name) != 0 { //we want a specific frontend
		var present bool
		outData, present = frontends[name]
		if !present {
			res.WriteHeader(http.StatusNotFound)
			//fmt.Fprint(res, string("Frontend not found"))
			return
		}

	} else { //if a name is not specified we return all the frontends
		outData = frontends
	}
	outgoingJSON, error := json.Marshal(outData)
	if error != nil {
		log.Println(error.Error())
		http.Error(res, error.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(res, string(outgoingJSON))
}

func newFrontend(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(req)
	name := vars["name"]
	_, present := frontends[name]
	if present {
		err := errors.Errorf("Frontend %s already present\n", name)
		log.Println(err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	newFront := nlb.Frontend{}
	decoder := json.NewDecoder(req.Body)
	error := decoder.Decode(&newFront)
	if error != nil {
		log.Println(error.Error())
		http.Error(res, error.Error(), http.StatusInternalServerError)
		return
	}
	now := time.Now()
	newFront.Metadata.CreatedAt = now
	newFront.Metadata.UpdatedAt = now
	frontends[newFront.Metadata.Name] = newFront
	outgoingJSON, err := json.Marshal(newFront.Metadata)
	if err != nil {
		log.Println(error.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusCreated)
	fmt.Fprint(res, string(outgoingJSON))
}

func editFrontends(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(req)
	name := vars["name"]

	switch req.Method {
	case "PUT":
		delete(frontends, name)
		newFrontend(res, req)
	case "DELETE":
		delete(frontends, name)
		res.WriteHeader(http.StatusNoContent)
	case "PATCH":
		editingFront, ok := frontends[name]
		if !ok {
			res.WriteHeader(http.StatusNotFound)
			//fmt.Fprint(res, string("Frontend not found"))
			return
		}

		frontend := nlb.Frontend{}
		decoder := json.NewDecoder(req.Body)
		error := decoder.Decode(&frontend)
		if error != nil {
			log.Println(error.Error())
			http.Error(res, error.Error(), http.StatusInternalServerError)
			return
		}
		frontend.Metadata.CreatedAt = editingFront.Metadata.CreatedAt
		frontend.Metadata.UpdatedAt = time.Now()
		frontends[name] = frontend
		outgoingJSON, err := json.Marshal(frontend)
		if err != nil {
			log.Println(error.Error())
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusCreated)
		fmt.Fprint(res, string(outgoingJSON))
	}
}

func listServices(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(req)
	name := vars["name"]

	var outData interface{}
	if len(name) != 0 { //we want a specific frontend
		var present bool
		outData, present = services[name]
		if !present {
			res.WriteHeader(http.StatusNotFound)
			//fmt.Fprint(res, string("Frontend not found"))
			return
		}

	} else { //if a name is not specified we return all the frontends
		outData = services
	}
	outgoingJSON, error := json.Marshal(outData)
	if error != nil {
		log.Println(error.Error())
		http.Error(res, error.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(res, string(outgoingJSON))
}

func newService(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(req)
	name := vars["name"]
	_, present := services[name]
	if present {
		err := errors.Errorf("Frontend %s already present\n", name)
		log.Println(err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	newSvc := nlb.Service{}
	decoder := json.NewDecoder(req.Body)
	error := decoder.Decode(&newSvc)
	if error != nil {
		log.Println(error.Error())
		http.Error(res, error.Error(), http.StatusInternalServerError)
		return
	}
	now := time.Now()
	newSvc.Metadata.CreatedAt = now
	newSvc.Metadata.UpdatedAt = now
	services[newSvc.Metadata.Name] = newSvc
	outgoingJSON, err := json.Marshal(newSvc.Metadata)
	if err != nil {
		log.Println(error.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusCreated)
	fmt.Fprint(res, string(outgoingJSON))
}

func editService(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(req)
	name := vars["name"]

	switch req.Method {
	case "PUT":
		delete(services, name)
		newFrontend(res, req)
	case "DELETE":
		delete(services, name)
		res.WriteHeader(http.StatusNoContent)
	case "PATCH":
		editingFront, ok := services[name]
		if !ok {
			res.WriteHeader(http.StatusNotFound)
			//fmt.Fprint(res, string("Frontend not found"))
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
		svc.Metadata.CreatedAt = editingFront.Metadata.CreatedAt
		svc.Metadata.UpdatedAt = time.Now()
		services[name] = svc
		outgoingJSON, err := json.Marshal(svc)
		if err != nil {
			log.Println(error.Error())
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusCreated)
		fmt.Fprint(res, string(outgoingJSON))
	}
}
