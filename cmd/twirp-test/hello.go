package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	cmd "github.com/jonabc/test-repo/gss/command"
	"github.com/jonabc/test-repo/gss/twirp/proto"
)

func helloCommand(c *cmd.Command) error {
	flagSet := flag.NewFlagSet("hello", flag.ExitOnError)
	serviceURLFlag := newServiceURLFlag(flagSet)

	if len(os.Args) < 3 {
		return cmd.NewUsageError("usage: twirp-test hello <name> [additional options]\n", flagSet)
	}

	name := os.Args[2]
	err := flagSet.Parse(os.Args[3:])
	if err != nil {
		return err
	}

	service := proto.NewHelloWorldAPIProtobufClient(*serviceURLFlag, &http.Client{})

	request := &proto.NameRequest{
		Name: name,
	}

	fmt.Printf("Calling Twirp server at %s with name %s...\n", *serviceURLFlag, name)

	response, err := service.HelloName(context.Background(), request)
	if err != nil {
		return err
	}

	fmt.Printf("Received response from server: %s\n", response.Message)

	return nil
}
