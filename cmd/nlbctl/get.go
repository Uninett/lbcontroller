package main

import (
	"fmt"

	"github.com/koki/json"

	"github.com/folago/nlb"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// getCmd represents the list command
var getCmd = &cobra.Command{
	Use:       "get RESOURCE [NAME]",
	Short:     "Display one or many resources.",
	Long:      "Display one or many resources.",
	Args:      cobra.MinimumNArgs(1),
	ValidArgs: []string{"service", "frontend"},
	//ArgAliases: []string{"service", "svc", "frontend", "frt"},
	RunE: getAction,
}

//TODO less copy paste?
func getAction(cmd *cobra.Command, args []string) error {
	var (
		resource = args[0]
	)

	if len(args) > 1 {
		name := args[1]
		switch resource {
		case "frontend":
			ret, err := nlb.GetFrontend(name, viper.GetString("url")+"/frontends")
			if err != nil {
				return errors.Wrap(err, "error getting resources")
			}
			data, err := json.MarshalIndent(ret, "", "  ")
			if err != nil {
				return errors.Wrap(err, "error printing response")
			}
			fmt.Printf("%s\n", data)
		case "service":
			fmt.Println("not implemented")
		default:
			fmt.Println("NOT A RESOURCE")
		}
	} else {
		switch resource {
		case "frontend":
			ret, err := nlb.ListFrontends(viper.GetString("url") + "/frontends")
			if err != nil {
				return errors.Wrap(err, "error getting resources")
			}

			data, err := json.MarshalIndent(ret, "", "  ")
			if err != nil {
				return errors.Wrap(err, "error printing response")
			}
			fmt.Printf("%s\n", data)
		case "service":
			fmt.Println("not implemented")
		default:
			fmt.Println("NOT A RESOURCE")
		}
	}
	return nil
}
