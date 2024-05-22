package main

import (
	"errors"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

func main() {
	var example string

	app := &cli.App{
		Name:  "examples",
		Usage: "run an example simulation",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "example",
				Aliases:     []string{"e"},
				Usage:       "Name of example",
				Required:    true,
				Destination: &example,
			},
		},
		Action: func(cCtx *cli.Context) error {
			return runExample(strings.ToLower(example))
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}

func runExample(example string) error {

	switch example {
	case "simple":
		simpleExample()
	case "vadere":
		vaderExample()
	case "high_traffic":
		highTrafficExample()
	case "sync":
		syncExample()
	default:
		return errors.New("given example name does not exist")
	}

	return nil
}
