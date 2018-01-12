package nlb

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/koki/json"
	"github.com/pkg/errors"
)

// ServiceConfig holds one of the different configurations
// of the services
type ServiceConfig interface {
	Type() ServiceType
}

//Service handled by the load balancers
type Service struct {
	Type     ServiceType   `json:"type,omitempty"`
	Metadata Metadata      `json:"metadata,omitempty"`
	Config   ServiceConfig `json:"config,omitempty"`
}

//ListServices return a list of services
//configured on the loadbalancers. Thee services
//are returned as Messages, with the actual configurartion
//still in json format.
func ListServices(url string) ([]Service, error) {
	res, err := http.Get(url + "/services")
	if err != nil {
		return nil, errors.Wrapf(err, "error connecting to API endpoint: %s\n ", url)
	}

	dec := json.NewDecoder(res.Body)
	svcs := []Service{}

	//read all the Messages and alter parse the cofigs
	for dec.More() {
		var s Service
		// decode an array value (Message)
		err := dec.Decode(&s)
		if err != nil {
			return nil, errors.Wrap(err, "error decoding a Service object")
		}
		svcs = append(svcs, s)
	}
	res.Body.Close()

	return svcs, nil
}

//MessageList is a list of Message that collcet some
//utility methods such as filtering by name or type.
type MessageList struct {
	List []Message
}

//ByType filter messages by type
func (l MessageList) ByType(typ ServiceType) []Message {
	var ret []Message
	for _, m := range l.List {
		if m.Type == typ {
			ret = append(ret, m)
		}
	}
	return ret
}

//NewService create a new service
func NewService(svc Service, url string) (*Metadata, error) {
	data, err := json.Marshal(svc)
	if err != nil {
		return nil, errors.Wrap(err, "error marshalling Service")
	}
	buf := bytes.NewBuffer(data)

	res, err := http.Post(url, jsonContent, buf)
	if err != nil {
		return nil, errors.Wrap(err, "error creating new service")
	}

	defer res.Body.Close()
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading from API endpoint: %s\n ", url)
	}
	ret := &Metadata{}
	err = json.Unmarshal(bytes, ret)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding frontend object")
	}
	return ret, nil
}
