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
	url = frontURL(url)

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
func GetFrontend(name, url string) (*Frontend, bool, error) {
	url = frontURL(url)
	ret := &Frontend{}

	res, err := http.Get(url + "/" + name)
	if err != nil {
		return nil, false, errors.Wrapf(err, "error connecting to API endpoint: %s\n ", url)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, false, errors.Wrapf(err, "error reading from API endpoint: %s\n ", url)
	}

	//handle stautus 200 and 404
	switch res.StatusCode {
	case http.StatusNotFound:
		return &Frontend{}, false, nil
	case http.StatusOK:
		break
	default:
		return nil, false, errors.Errorf("error, returned status from API endpoint not supported: %s\n ", res.Status)
	}

	err = json.Unmarshal(body, ret)
	if err != nil {
		return nil, false, errors.Errorf("API endpoint returned status %s, %s\n ", res.Status, bytes.TrimSpace(body))
	}

	return ret, true, nil
}

//NewFrontend creates a new frontend, the new frontent created is returned.
//Currently, only frontend objects of type address are supported.
//They allow for specific control of which addresses are used
//(and possibly reused) by the various services.
//If this control is not needed, frontend objects will be created
//automaticly if needed on service creation.
func NewFrontend(front Frontend, url string) (*Metadata, error) {
	url = frontURL(url)
	data, err := json.Marshal(front)
	if err != nil {
		return nil, errors.Wrap(err, "error marshalling Frontend")
	}
	buf := bytes.NewBuffer(data)

	res, err := http.Post(url, jsonContent, buf)
	if err != nil {
		return nil, errors.Wrap(err, "error POSTing new frontend")
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading from API endpoint: %s", url)
	}

	if res.StatusCode != http.StatusCreated {
		return nil, errors.Errorf("API endpoint returned status %s, %s", res.Status, bytes.TrimSpace(body))
	}

	ret := &Metadata{}
	err = json.Unmarshal(body, ret)
	if err != nil {
		return nil, errors.Wrapf(err, "error decoding Metadata object %s", body)
	}
	return ret, nil
}

//ReplaceFrontend replace and exixting frontend object, the new Frontend is retured.
func ReplaceFrontend(front Frontend, url string) (*Frontend, error) {
	url = frontURL(url)

	req, err := prepareRequest(front, url+"/"+front.Metadata.Name, "PUT")
	if err != nil {
		return nil, errors.Wrap(err, "error creatign http.Request")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "error replacing frontend %s/n", front.Metadata.Name)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading from API endpoint: %s", url)
	}
	ret := &Frontend{}
	err = json.Unmarshal(body, ret)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding frontend object")
	}
	return ret, nil
}

//ReconfigFrontend replace and exixting frontend object, the new Frontend is retured.
func ReconfigFrontend(front Frontend, url string) (*Frontend, error) {
	url = frontURL(url)

	req, err := prepareRequest(front, url+"/"+front.Metadata.Name, "PATCH")
	if err != nil {
		return nil, errors.Wrap(err, "error creatign http.Request")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "error reconfiguring frontend %s/n", front.Metadata.Name)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading from API endpoint: %s", url)
	}
	ret := &Frontend{}
	err = json.Unmarshal(body, ret)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding frontend object")
	}
	return ret, nil
}

//DeleteFrontend replace and exixting frontend object, the new Frontend is retured.
func DeleteFrontend(name, url string) error {
	url = frontURL(url)

	req, err := http.NewRequest("DELETE", url+"/"+name, nil)
	if err != nil {
		return errors.Wrap(err, "error creatign http.Request")
	}
	req.Header.Set("Content-Type", jsonContent)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "error replacing frontend %s", name)
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

//EditFrontend edit and existing frontend object according to the action, the new Frontend is retured.
func EditFrontend(front Frontend, url string, action action) (*Frontend, error) {
	url = frontURL(url)
	var httpMethod string
	switch action {
	case replace:
		httpMethod = "PUT"
	case reconfig:
		httpMethod = "PATCH"
	case delete:
		httpMethod = "DELETE"
	default:
		return nil, errors.Errorf("unrecognized action %s", action)
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
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading from API endpoint: %s", url)
	}
	ret := &Frontend{}
	err = json.Unmarshal(body, ret)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding frontend object")
	}
	return ret, nil
}

func frontURL(url string) string {
	return url + "/" + "frontends"
}
