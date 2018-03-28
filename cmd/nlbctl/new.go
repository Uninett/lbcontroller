package main

import (
	"fmt"
	"io"
	"os"

	"github.com/folago/nlb"
	"github.com/koki/json"
	"github.com/pkg/errors"
	"gopkg.in/urfave/cli.v1"
)

func newService(c *cli.Context) error {

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
	svc := &nlb.Service{}
	err = dec.Decode(svc)
	if err != nil {
		return errors.Wrap(err, "error decoding json resource file")
	}
	ingress, err := nlb.NewService(*svc, apiURL)
	if err != nil {
		return errors.Wrap(err, "error creating new service")
	}
	fmt.Printf("service created, ingress: %v\n", ingress)

	return nil
}
