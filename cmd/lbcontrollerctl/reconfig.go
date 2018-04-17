package main

import (
	"fmt"
	"io"
	"os"

	"github.com/koki/json"
	"github.com/pkg/errors"
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
	err = lbcontroller.ReconfigService(*svc, apiURL)
	if err != nil {
		return errors.Wrap(err, "error configuring service")
	}
	fmt.Printf("service %s reconfigured\n", svc.Metadata.Name)

	return nil
}

func reconfigFrontend(c *cli.Context) error {
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
	fnt := &lbcontroller.Frontend{}
	err = dec.Decode(fnt)
	if err != nil {
		return errors.Wrap(err, "error decoding json resource file")
	}
	err = lbcontroller.ReconfigFrontend(*fnt, apiURL)
	if err != nil {
		return errors.Wrap(err, "error configuring frontend")
	}
	fmt.Printf("frontend %s reconfigured\n", fnt.Metadata.Name)

	return nil
}
