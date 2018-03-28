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
	svc := &nlb.Service{}
	err = dec.Decode(svc)
	if err != nil {
		return errors.Wrap(err, "error decoding json resource file")
	}
	err = nlb.ReconfigService(*svc, apiURL)
	if err != nil {
		return errors.Wrap(err, "error configuring service")
	}
	fmt.Printf("service %s reconfigured\n", svc.Metadata.Name)

	return nil
}
