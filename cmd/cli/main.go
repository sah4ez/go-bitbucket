package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/urfave/cli"

	sw "github.com/sah4ez/go-bitbucket"
)

var (
	Revision = ""
	Version  = ""
)

var (
	Auth    context.Context
	Repo    string
	Company string
	PrID    int64
)

func init() {
	cred := os.Getenv("BITBUCKET_CREDENTIAL")
	parts := strings.Split(cred, ":")
	Auth = context.WithValue(context.Background(), sw.ContextBasicAuth, sw.BasicAuth{
		UserName: parts[0],
		Password: parts[1],
	})

	Repo = os.Getenv("REPO")
	Company = os.Getenv("COMPANY")
	PrID, _ = strconv.ParseInt(os.Getenv("PR"), 10, 32)
}

func main() {
	config := sw.NewConfiguration()

	client := sw.NewAPIClient(config)

	app := cli.NewApp()
	app.Name = "comments"
	app.Version = Version

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name: "debug, d",
		},
		cli.StringFlag{
			Name:   "prefix, p",
			Value:  "/tmp",
			EnvVar: "PREFIX",
		},
	}
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("version: %s\nrevision: %s\n", c.App.Version, Revision)
	}

	app.Commands = []cli.Command{
		Review(client),
		Comments(client),
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println("Couldn't start application: ", err.Error())
		os.Exit(2)
	}

}
