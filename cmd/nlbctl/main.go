// Copyright © 2018 Giacomo Tartari
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice,
//    this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation
//    and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its contributors
//    may be used to endorse or promote products derived from this software
//    without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package main

import (
	"log"
	"os"

	"gopkg.in/urfave/cli.v1"
)

var (
	cfgFile string
	apiURL  string
	resFile string
)

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "api",
			Value:       "http://127.0.0.1:8080",
			Usage:       "API endpoint",
			Destination: &apiURL,
			EnvVar:      "NLB_API",
		},
	}

	app.Commands = []cli.Command{
		{
			Name: "get",
			//Aliases: []string{"t"},
			Usage: "Display one or many resources",
			Subcommands: []cli.Command{
				{
					Name:    "frontend",
					Usage:   "display frontend(s)",
					Aliases: []string{"fnt"},
					Action:  getFrontend,
				},
				{
					Name:    "service",
					Usage:   "display service(s)",
					Aliases: []string{"svc"},
					Action:  getService,
				},
			},
		},
		{
			Name:    "new",
			Aliases: []string{"create"},
			Usage:   " create new resources",
			Subcommands: []cli.Command{
				{
					Name:      "frontend",
					Usage:     "create a new frontend",
					UsageText: "create a new frontend, if a file is specified witht the --file(-f) flag, the file is read. Oterwise stdin is used.",
					Aliases:   []string{"fnt"},
					Action:    newFrontend,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:        "file, f",
							Value:       "",
							Usage:       "json file containing the resourse description",
							Destination: &resFile,
						},
					},
				},
				{
					Name:      "service",
					Usage:     "create a new service",
					UsageText: "create a new service, if a file is specified witht the --file(-f) flag, the file is read. Oterwise stdin is used.",
					Aliases:   []string{"svc"},
					Action:    newService,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:        "file, f",
							Value:       "",
							Usage:       "json file containing the resourse description",
							Destination: &resFile,
						},
					},
				},
			},
		},
		{
			Name:    "delete",
			Aliases: []string{"rm", "del"},
			Usage:   "Delete a resource",
			Subcommands: []cli.Command{
				{
					Name:    "frontend",
					Usage:   "delete a frontend",
					Aliases: []string{"fnt"},
					Action:  delFrontend,
				},
				{
					Name:    "service",
					Usage:   "delete a service",
					Aliases: []string{"svc"},
					Action:  delService,
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
