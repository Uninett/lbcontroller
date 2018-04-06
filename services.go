package nlb

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/koki/json"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
)

//Service handled by the load balancers
type Service struct {
	Type     ServiceType              `json:"type,omitempty"`
	Metadata Metadata                 `json:"metadata,omitempty"`
	Config   Config                   `json:"config,omitempty"`
	Ingress  []v1.LoadBalancerIngress `json:"ingress,omitempty"`
}

//ListServices return a list of services
//configured on the loadbalancers. Thee services
//are returned as Messages, with the actual configurartion
//still in json format.
func ListServices(url string) ([]Service, error) {
	url = svcURL(url)

	res, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "error connecting to API endpoint: %s", url)
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.Errorf("error, returned status not 200 OK from API endpoint: %s", res.Status)
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

//GetService get the configuration of the fronten specified by name, if the service
//is found GetService returnns a true boolean value as well
func GetService(name, url string) (*Service, bool, error) {

	url = svcURL(url)
	ret := &Service{}

	res, err := http.Get(url + "/" + name)
	if err != nil {
		return nil, false, errors.Wrapf(err, "error connecting to API endpoint: %s", url)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, false, errors.Wrapf(err, "error reading from API endpoint: %s", url)
	}

	var ingress []v1.LoadBalancerIngress
	//handle stautus 200 and 404
	switch res.StatusCode {
	case http.StatusNotFound:
		return &Service{}, false, nil
	case http.StatusOK:
		location := res.Header.Get("Location")
		if location != "" {
			ingress, err = GetIngress(location)
			if err != nil {
				return nil, false, errors.Wrapf(err, "error getting ingress form api: %s", location)
			}
		}

	default:
		return nil, false, errors.Errorf("error, returned status from API endpoint not supported: %s\n ", res.Status)
	}

	err = json.Unmarshal(body, ret)
	if err != nil {
		return nil, false, errors.Wrap(err, "error decoding Service object")
	}
	ret.Ingress = ingress

	return ret, true, nil
}

//NewService create a new service
func NewService(svc Service, url string) ([]v1.LoadBalancerIngress, error) {

	url = svcURL(url)
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

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading from API endpoint: %s", url)
	}

	if res.StatusCode != http.StatusCreated {
		return nil, errors.Errorf("API endpoint returned status %s, %s", res.Status, body)
	}

	location := res.Header.Get("Location")
	var ret []v1.LoadBalancerIngress

	if location != "" {
		ret, err = GetIngress(location)
		if err != nil {
			return nil, errors.Wrapf(err, "error getting ingress form api: %s", location)
		}
		return ret, nil
	}

	err = json.Unmarshal(body, &ret)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding Service object")
	}
	return ret, nil
}

//ReconfigService replace and exixting Service object, the new Service is retured.
func ReconfigService(svc Service, url string) ([]v1.LoadBalancerIngress, error) {
	url = svcURL(url)

	req, err := prepareRequest(svc, url+"/"+svc.Metadata.Name, "PATCH")
	if err != nil {
		return nil, errors.Wrap(err, "error creatign http.Request")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "error reconfiguring Service %s", svc.Metadata.Name)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading from API endpoint: %s", url)
	}

	var ret []v1.LoadBalancerIngress

	switch res.StatusCode {

	case http.StatusCreated:
		err = json.Unmarshal(body, &ret)
		if err != nil {
			return nil, errors.Wrap(err, "error decoding LoadBalancerIngress object")
		}

	case http.StatusNoContent:
		location := res.Header.Get("Location")
		if location != "" {
			ret, err = GetIngress(location)
			if err != nil {
				return nil, errors.Wrapf(err, "error getting ingress form api: %s", location)
			}
			return ret, nil
		}

	default:
		return nil, errors.Errorf("API endpoint returned status %s, %s", res.Status, bytes.TrimSpace(body))
	}

	return ret, nil
}

//DeleteService deletes and exixting Service object, the new Service is retured.
func DeleteService(name, url string) error {
	url = svcURL(url)

	req, err := http.NewRequest("DELETE", url+"/"+name, nil)
	if err != nil {
		return errors.Wrap(err, "error creatign http.Request")
	}
	req.Header.Set("Content-Type", jsonContent)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "error replacing Service %s", name)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return errors.Wrapf(err, "error reading from API endpoint: %s", url)
		}
		return errors.Errorf("API endpoint returned status %s, %s", res.Status, bytes.TrimSpace(body))
	}

	return nil
}

func svcURL(url string) string {
	return url + "/" + servicePath
}

//GetIngress retrives the k8s loadBalancerIngress from the specified url
func GetIngress(url string) ([]v1.LoadBalancerIngress, error) {

	//url = svcURL(url)
	var ret []v1.LoadBalancerIngress

	res, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "error connecting to API endpoint to get ingress: %s", url)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading from API endpoint to get ingress: %s", url)
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.Errorf("error, returned status from API endpoint not supported: %s\n ", res.Status)
	}

	err = json.Unmarshal(body, &ret)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding LoadBalancerIngress object")
	}

	return ret, nil
}
