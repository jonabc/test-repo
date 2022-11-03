package main

import (
	"fmt"
	"log"
	"os"
	"sort"

	_ "github.com/go-sql-driver/mysql"
	"github.com/urfave/cli/v2"
)

func main() {
	if err := realMain(); err != nil {
		log.Fatal(err)
	}
}

func realMain() error {

	app := &cli.App{
		Name:  "transition",
		Usage: "Run transitions, or GHE migrations and transitions with this small CLI",
		Commands: []*cli.Command{
			{
				Name:    "enterprise",
				Aliases: []string{"e"},
				Usage:   "Run database migrations and transitions in GHE",
				Action:  EnterpriseAction,
			},
			{
				Name:    "cloud",
				Aliases: []string{"c"},
				Usage:   "Run transitions only for cloud environments",
				Action:  CloudAction,
			},
		},
	}

	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		return fmt.Errorf("running CLI: %v", err)
	}

	return nil
}

// EnterpriseAction runs migrations and transitions together, for use in GHE environments.
func EnterpriseAction(c *cli.Context) error {
	// Your application will likely have its own configuration, app context, and database helper code to plug in here.
	return nil
}

// CloudAction runs the data transitions only, for use in cloud environments (locally and with Moda jobs).
func CloudAction(c *cli.Context) error {
	// Your application will likely have its own configuration, app context, and database helper code to plug in here
	return nil
}
