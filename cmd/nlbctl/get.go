package main

import (
	"fmt"
	"os"

	"text/tabwriter"

	"github.com/folago/nlb"
	"github.com/koki/json"
	"github.com/pkg/errors"
	"gopkg.in/urfave/cli.v1"
)

func getService(c *cli.Context) error {
	args := c.Args()

	if len(args) == 0 { //get all
		ret, err := nlb.ListServices(apiURL)
		if err != nil {
			return errors.Wrap(err, "error getting resources")
		}
		printServiceList(ret)
	}

	if len(args) == 1 { //get the first
		name := args[0]
		ret, found, err := nlb.GetService(name, apiURL)
		if err != nil {
			return errors.Wrap(err, "error getting resources")
		}
		if !found {
			fmt.Printf("no service found found with name %s\n", name)
			return nil
		}
		data, err := json.MarshalIndent(ret, "", "  ")
		if err != nil {
			return errors.Wrap(err, "error printing response")
		}
		fmt.Printf("%s\n", data)
	}
	return nil
}

func printServiceList(servs []nlb.Service) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 8, 8, 1, '\t', 0)
	fmt.Fprintln(w, "SERVICES\tNAME\tTYPE")
	for _, sv := range servs {
		fmt.Fprintf(w, "\t%s\t%s\n", sv.Metadata.Name, sv.Type)
	}
	w.Flush()
}
