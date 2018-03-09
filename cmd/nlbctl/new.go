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
	meta, err := nlb.NewService(*svc, apiURL)
	if err != nil {
		return errors.Wrap(err, "error creating new service")
	}
	fmt.Printf("service %s, created at %v\n", meta.Name, meta.CreatedAt)

	return nil
}

func newFrontend(c *cli.Context) error {
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
	fnt := &nlb.Frontend{}
	err = dec.Decode(fnt)
	if err != nil {
		return errors.Wrap(err, "error decoding json resource file")
	}
	meta, err := nlb.NewFrontend(*fnt, apiURL)
	if err != nil {
		return errors.Wrap(err, "error creating new frontend")
	}
	fmt.Printf("frontend %s, created at %v\n", meta.Name, meta.CreatedAt)

	return nil
}
