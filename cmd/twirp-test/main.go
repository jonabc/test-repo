package main

import (
	"errors"
	"flag"
	"log"
	"os"

	cmd "github.com/jonabc/test-repo/gss/command"
)

var allCommands = []cmd.Command{
	{Name: "hello", Desc: "Interacts with the hello world example service", Execute: helloCommand},
	{Name: "banner", Desc: "Interacts with the GitHub banner example service", Execute: bannerCommand, Commands: []cmd.Command{
		{Name: "add", Desc: "Add a new banner", Execute: bannerAddCommand},
		{Name: "update", Desc: "Update an existing banner", Execute: bannerUpdateCommand},
		{Name: "remove", Desc: "Remove a banner", Execute: bannerRemoveCommand},
		{Name: "get", Desc: "Retrieve info for a specific banner", Execute: bannerGetCommand},
		{Name: "list", Desc: "List banners", Execute: bannerListCommand},
	}},
}

func newServiceURLFlag(f *flag.FlagSet) *string {
	return f.String("url", "http://localhost:8080", "The url of the Twirp server")
}

func main() {
	if err := realMain(); err != nil {
		log.Fatalf("failed to run service: %v", err)
	}
}

func realMain() error {
	if len(os.Args) < 2 {
		return errors.New("usage: twirp-test <command> [additional options]\n" + cmd.ListCommands(allCommands))
	}

	return cmd.DispatchCommand(os.Args[1], allCommands)
}
