package main

import "github.com/pkg/errors"

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
		err := lbcontroller.DeleteService(name, apiURL)
		if err != nil {
			return errors.Wrapf(err, "error deleting service %s", name)
		}
	}
	return nil
}

func delFrontend(c *cli.Context) error {
	args := c.Args()

	if len(args) == 0 { //not enough
		return errors.New("Please specify a frontend to delete")
	}
	if len(args) > 1 { //too many
		return errors.New("Please specify only one frontend to delete")
	}

	if len(args) == 1 {
		name := args[0]
		err := lbcontroller.DeleteFrontend(name, apiURL)
		if err != nil {
			return errors.Wrapf(err, "error deleting frontend %s", name)
		}
	}
	return nil
}
