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

// UnmarshalJSON implements json.Unmarshaller interface
func (svc *Service) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		return nil
	}
	msg := new(Message)
	err := json.Unmarshal(data, msg)
	if err != nil {
		return errors.Wrap(err, "Error unmarsalling service")
	}

	switch msg.Type {
	case TCP, TCPProxyProtocol:
		tcpConf := new(TCPConfig)
		err = json.Unmarshal(msg.Config, tcpConf)
		if err != nil {
			return errors.Wrap(err, "Error unmarsalling TCP service configuration")
		}
		svc.Type = msg.Type
		svc.Metadata = msg.Metadata
		svc.Config = tcpConf
	case SharedHTTP:
		httpConf := new(SharedHTTPConfig)
		err = json.Unmarshal(msg.Config, httpConf)
		if err != nil {
			return errors.Wrap(err, "Error unmarsalling SharedHTTP service configuration")
		}
		svc.Type = msg.Type
		svc.Metadata = msg.Metadata
		svc.Config = httpConf
	case Mediasite:
		return errors.New("mediasite service unsupported")
	}

	return nil
}

//ListServices return a list of services
//configured on the loadbalancers. Thee services
//are returned as Messages, with the actual configurartion
//still in json format.
func ListServices(url string) ([]Service, error) {
	res, err := http.Get(url)
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

//NewService create a new service
func NewService(svc Service, url string) (*Metadata, error) {
	data, err := json.Marshal(svc)
	if err != nil {
		return nil, errors.Wrap(err, "error marshalling Service")
	}
	buf := bytes.NewBuffer(data)

	res, err := http.Post(url+"/"+svc.Metadata.Name, jsonContent, buf)
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
		return nil, errors.Wrap(err, "error decoding Service object")
	}
	return ret, nil
}

//GetService get the configuration of the fronten specified by name
func GetService(name, url string) (*Service, error) {
	ret := &Service{}

	res, err := http.Get(url + "/" + name)
	if err != nil {
		return nil, errors.Wrapf(err, "error connecting to API endpoint: %s\n ", url)
	}
	defer res.Body.Close()

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading from API endpoint: %s\n ", url)
	}

	err = json.Unmarshal(bytes, ret)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding Service object")
	}

	return ret, nil
}

//ReplaceService replace and exixting Service object, the new Service is retured.
func ReplaceService(front Service, url string) (*Service, error) {

	req, err := prepareRequest(front, url+"/"+front.Metadata.Name, "PUT")
	if err != nil {
		return nil, errors.Wrap(err, "error creatign http.Request")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "error replacing Service %s/n", front.Metadata.Name)
	}

	defer res.Body.Close()
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading from API endpoint: %s\n ", url)
	}
	ret := &Service{}
	err = json.Unmarshal(bytes, ret)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding Service object")
	}
	return ret, nil
}

//ReconfigService replace and exixting Service object, the new Service is retured.
func ReconfigService(front Service, url string) (*Service, error) {

	req, err := prepareRequest(front, url+"/"+front.Metadata.Name, "PATCH")
	if err != nil {
		return nil, errors.Wrap(err, "error creatign http.Request")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "error reconfiguring Service %s/n", front.Metadata.Name)
	}

	defer res.Body.Close()
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading from API endpoint: %s\n ", url)
	}
	ret := &Service{}
	err = json.Unmarshal(bytes, ret)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding Service object")
	}
	return ret, nil
}

//DeleteService replace and exixting Service object, the new Service is retured.
func DeleteService(name, url string) (*Service, error) {

	req, err := http.NewRequest("PUT", url+"/"+name, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creatign http.Request")
	}
	req.Header.Set("Content-Type", jsonContent)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "error replacing Service %s/n", name)
	}

	defer res.Body.Close()
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading from API endpoint: %s\n ", url)
	}
	ret := &Service{}
	err = json.Unmarshal(bytes, ret)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding Service object")
	}
	return ret, nil
}
