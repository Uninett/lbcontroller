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
	frontends = map[string]nlb.Frontend{}
	services  = map[string]nlb.Service{}
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/frontends", listFrontends).Methods("GET")
	router.HandleFunc("/frontends/{name}", getFrontend).Methods("GET")
	router.HandleFunc("/frontends", newFrontend).Methods("POST")
	router.HandleFunc("/frontends/{name}", editFrontends).Methods("PUT", "DELETE", "PATCH")

	router.HandleFunc("/services", listServices).Methods("GET")
	router.HandleFunc("/services/{name}", getService).Methods("GET")
	router.HandleFunc("/services", newService).Methods("POST")
	router.HandleFunc("/services/{name}", editService).Methods("PUT", "DELETE", "PATCH")

	loggedRouter := handlers.LoggingHandler(os.Stdout, router)
	http.ListenAndServe("127.0.0.1:8080", loggedRouter)
}

func getFrontend(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(req)
	name := vars["name"]

	if len(name) == 0 {
		log.Println("need to specify a resource name")
		http.Error(res, "need to specify a resource name", http.StatusInternalServerError)
		return
	}

	outData, present := frontends[name]
	if !present {
		res.WriteHeader(http.StatusNotFound)
		//fmt.Fprint(res, string("Frontend not found"))
		return
	}
	outgoingJSON, error := json.Marshal(outData)
	if error != nil {
		log.Println(error.Error())
		http.Error(res, error.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(res, string(outgoingJSON))
}

func listFrontends(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	for _, frt := range frontends {

		outgoingJSON, error := json.Marshal(frt)
		if error != nil {
			log.Println(error.Error())
			http.Error(res, error.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(res, string(outgoingJSON))
	}
}

func newFrontend(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")

	newFront := nlb.Frontend{}
	decoder := json.NewDecoder(req.Body)
	error := decoder.Decode(&newFront)
	if error != nil {
		log.Println(error.Error())
		http.Error(res, error.Error(), http.StatusInternalServerError)
		return
	}
	_, present := frontends[newFront.Metadata.Name]
	if present {
		err := errors.Errorf("Frontend %s already present\n", newFront.Metadata.Name)
		log.Println(err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
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
		res.WriteHeader(http.StatusNoContent)
	case "DELETE":
		delete(frontends, name)
		res.WriteHeader(http.StatusNoContent)
	case "PATCH":
		editingFront, ok := frontends[name]
		if !ok {
			res.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(res, "Frontend %s not found", name)
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
		//outgoingJSON, err := json.Marshal(frontend)
		//if err != nil {
		//	log.Println(error.Error())
		//	http.Error(res, err.Error(), http.StatusInternalServerError)
		//	return
		//}
		//res.WriteHeader(http.StatusCreated)
		//fmt.Fprint(res, string(outgoingJSON))
		res.WriteHeader(http.StatusNoContent)
	}
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
		_, ok := services[name]
		if !ok {
			res.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(res, "Service %s not found", name)
			return
		}
		delete(services, name)
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
		services[name] = newSvc
		res.WriteHeader(http.StatusNoContent)
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
		//outgoingJSON, err := json.Marshal(svc)
		//if err != nil {
		//	log.Println(error.Error())
		//	http.Error(res, err.Error(), http.StatusInternalServerError)
		//	return
		//}
		//res.WriteHeader(http.StatusCreated)
		//fmt.Fprint(res, string(outgoingJSON))
		res.WriteHeader(http.StatusNoContent)
	}
}
