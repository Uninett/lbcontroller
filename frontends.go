package nlb

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/koki/json"
	"github.com/pkg/errors"
)

//ListFrontends get a list of frontends configured on the load balancers
func ListFrontends(url string) ([]Frontend, error) {

	res, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "error connecting to API endpoint: %s\n ", url)
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.Errorf("error, returned status not 200 OK from API endpoint: %s\n ", res.Status)
	}

	dec := json.NewDecoder(res.Body)
	msgs := []Frontend{}

	//read all the Messages and parse the cofigs
	for dec.More() {
		var m Frontend
		// decode an array value (Message)
		err := dec.Decode(&m)
		if err != nil {
			return nil, errors.Wrap(err, "error decoding a frontend object")
		}
		if m.Type != "frontend" {
			return nil, errors.Errorf("expected a frontend got %v\n", m)
		}
		msgs = append(msgs, m)
	}
	res.Body.Close()

	return msgs, nil
}

//GetFrontend get the configuration of the fronten specified by name
func GetFrontend(name, url string) (*Frontend, error) {
	ret := &Frontend{}

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
		return nil, errors.Wrap(err, "error decoding frontend object")
	}

	return ret, nil
}

//NewFrontend creates a new frontend, the new frontent created is returned.
//Currently, only frontend objects of type address are supported.
//They allow for specific control of which addresses are used
//(and possibly reused) by the various services.
//If this control is not needed, frontend objects will be created
//automaticly if needed on service creation.
func NewFrontend(front Frontend, url string) (*Frontend, error) {
	data, err := json.Marshal(front)
	if err != nil {
		return nil, errors.Wrap(err, "error marshalling Frontend")
	}
	buf := bytes.NewBuffer(data)

	res, err := http.Post(url, jsonContent, buf)
	if err != nil {
		return nil, errors.Wrap(err, "error POSTing new frontend")
	}
	// now we have to return the new frontend

	defer res.Body.Close()
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading from API endpoint: %s\n ", url)
	}
	ret := &Frontend{}
	err = json.Unmarshal(bytes, ret)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding frontend object")
	}
	return ret, nil
}

//ReplaceFrontend replace and exixting frontend object, the new Frontend is retured.
func ReplaceFrontend(front Frontend, url string) (*Frontend, error) {

	req, err := prepareRequest(front, url+"/"+front.Metadata.Name, "PUT")
	if err != nil {
		return nil, errors.Wrap(err, "error creatign http.Request")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "error replacing frontend %s/n", front.Metadata.Name)
	}

	defer res.Body.Close()
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading from API endpoint: %s\n ", url)
	}
	ret := &Frontend{}
	err = json.Unmarshal(bytes, ret)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding frontend object")
	}
	return ret, nil
}

//ReconfigFrontend replace and exixting frontend object, the new Frontend is retured.
func ReconfigFrontend(front Frontend, url string) (*Frontend, error) {

	req, err := prepareRequest(front, url+"/"+front.Metadata.Name, "PATCH")
	if err != nil {
		return nil, errors.Wrap(err, "error creatign http.Request")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "error reconfiguring frontend %s/n", front.Metadata.Name)
	}

	defer res.Body.Close()
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading from API endpoint: %s\n ", url)
	}
	ret := &Frontend{}
	err = json.Unmarshal(bytes, ret)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding frontend object")
	}
	return ret, nil
}

//DeleteFrontend replace and exixting frontend object, the new Frontend is retured.
func DeleteFrontend(name, url string) (*Frontend, error) {

	req, err := http.NewRequest("PUT", url+"/"+name, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creatign http.Request")
	}
	req.Header.Set("Content-Type", jsonContent)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "error replacing frontend %s/n", name)
	}

	defer res.Body.Close()
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading from API endpoint: %s\n ", url)
	}
	ret := &Frontend{}
	err = json.Unmarshal(bytes, ret)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding frontend object")
	}
	return ret, nil
}

//EditFrontend edit and existing frontend object according to the action, the new Frontend is retured.
func EditFrontend(front Frontend, url string, action action) (*Frontend, error) {
	var httpMethod string
	switch action {
	case replace:
		httpMethod = "PUT"
	case reconfig:
		httpMethod = "PATCH"
	case delete:
		httpMethod = "DELETE"
	default:
		return nil, errors.Errorf("unrecognized action %s/n", action)
	}

	req, err := prepareRequest(front, url+"/"+front.Metadata.Name, httpMethod)
	if err != nil {
		return nil, errors.Wrap(err, "error creatign http.Request")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "error editing %s/n", front.Metadata.Name)
	}

	defer res.Body.Close()
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading from API endpoint: %s\n ", url)
	}
	ret := &Frontend{}
	err = json.Unmarshal(bytes, ret)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding frontend object")
	}
	return ret, nil
}
