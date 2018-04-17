package main

import (
	"fmt"
	"io"
	"os"

	"github.com/koki/json"
	"github.com/pkg/errors"
)

func replaceService(c *cli.Context) error {

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
	loc, err := lbcontroller.ReplaceService(*svc, apiURL)
	if err != nil {
		return errors.Wrap(err, "error configuring service")
	}
	fmt.Printf("service %s reconfigured: %s\n", svc.Metadata.Name, loc)

	return nil
}

func replaceFrontend(c *cli.Context) error {
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
	loc, err := lbcontroller.ReplaceFrontend(*fnt, apiURL)
	if err != nil {
		return errors.Wrap(err, "error configuring frontend")
	}
	fmt.Printf("frontend %s reconfigured: %s\n", fnt.Metadata.Name, loc)

	return nil
}
