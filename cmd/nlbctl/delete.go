package main

import (
	"github.com/folago/nlb"
	"github.com/pkg/errors"
	"gopkg.in/urfave/cli.v1"
)

func delService(c *cli.Context) error {
	args := c.Args()

	if len(args) == 0 { //not enough
		return errors.New("Please specify a service to delete")
	}
	if len(args) > 1 { //too many
		return errors.New("Please specify only one service to delete")
	}

	if len(args) == 1 {
		name := args[0]
		err := nlb.DeleteService(name, apiURL)
		if err != nil {
			return errors.Wrapf(err, "error deleting service %s", name)
		}
	}
	return nil
}
