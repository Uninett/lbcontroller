package main

import (
	"fmt"
	"io"
	"os"

	"github.com/koki/json"
	"github.com/pkg/errors"
	"github.com/uninett/lbcontroller"
)

func reconfigService(c *cli.Context) error {

	if len(c.Args()) != 0 {
		return errors.New("too many args")
	}
	var (
		indata io.Reader
		err    error
	)
	if resFile != "" {
		indata, err = os.Open(resFile)
		if err != nil {
			return errors.Wrap(err, "error opening resourse file")
		}
	} else {
		indata = os.Stdin
	}
	dec := json.NewDecoder(indata)
	svc := &lbcontroller.Service{}
	err = dec.Decode(svc)
	if err != nil {
		return errors.Wrap(err, "error decoding json resource file")
	}
	ingress, err := lbcontroller.ReconfigService(*svc, apiURL)
	if err != nil {
		return errors.Wrap(err, "error configuring service")
	}
	fmt.Printf("service %s reconfigured\n%v\n", svc.Metadata.Name, ingress)

	return nil
}
