package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	cmd "github.com/jonabc/test-repo/gss/command"
	"github.com/jonabc/test-repo/gss/twirp/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func bannerCommand(c *cmd.Command) error {
	if len(os.Args) < 3 {
		return errors.New("usage: twirp-test banner <command> [additional options]\n" + cmd.ListCommands(c.Commands))
	}

	return cmd.DispatchCommand(os.Args[2], c.Commands)
}

func bannerAddCommand(c *cmd.Command) error {
	flagSet := flag.NewFlagSet("banner add", flag.ExitOnError)
	typeFlag := flagSet.String("type", "info", "The type of banner (info, warning, error)")
	expiresFlag := flagSet.String("expires", "", "Date/time when the banner should no longer be presented to users")
	messageFlag := flagSet.String("message", "", "The message string to show in the banner")
	serviceURLFlag := newServiceURLFlag(flagSet)

	if len(os.Args) < 4 {
		return cmd.NewUsageError("usage: twirp-test banner add [options]\n", flagSet)
	}

	err := flagSet.Parse(os.Args[3:])
	if err != nil {
		return err
	}

	var bannerType proto.Banner_BannerType
	switch *typeFlag {
	case "info":
		bannerType = proto.Banner_Info
	case "warning":
		bannerType = proto.Banner_Warning
	case "error":
		bannerType = proto.Banner_Error
	default:
		return errors.New("the banner type is not recognized")
	}

	var timestamp *timestamppb.Timestamp
	if expiresFlag != nil && *expiresFlag != "" {
		t, err := time.Parse(time.RFC3339, *expiresFlag)
		if err != nil {
			return err
		}

		timestamp = timestamppb.New(t)
		if err != nil {
			return err
		}
	}

	err = flagSet.Parse(os.Args[3:])
	if err != nil {
		return err
	}

	service := proto.NewBannersAPIProtobufClient(*serviceURLFlag, &http.Client{})

	request := &proto.AddBannerRequest{
		Banner: &proto.Banner{
			BannerType: bannerType,
			ExpiresAt:  timestamp,
			Message:    *messageFlag,
		},
	}

	response, err := service.AddBanner(context.Background(), request)
	if err != nil {
		return err
	}

	json, err := json.Marshal(response.Banner)
	if err != nil {
		return err
	}

	fmt.Println(string(json))
	return nil
}

func bannerUpdateCommand(c *cmd.Command) error {
	flagSet := flag.NewFlagSet("banner update", flag.ExitOnError)
	idFlag := flagSet.Int("id", 0, "The banner id")
	expiresFlag := flagSet.String("expires", "", "Date/time when the banner should no longer be presented to users")
	messageFlag := flagSet.String("message", "", "The message string to show in the banner")
	serviceURLFlag := newServiceURLFlag(flagSet)

	if len(os.Args) < 4 {
		return cmd.NewUsageError("usage: twirp-test banner update [options]\n", flagSet)
	}

	err := flagSet.Parse(os.Args[3:])
	if err != nil {
		return err
	}

	var timestamp *timestamppb.Timestamp
	if expiresFlag != nil && *expiresFlag != "" {
		t, err := time.Parse(time.RFC3339, *expiresFlag)
		if err != nil {
			return err
		}

		timestamp = timestamppb.New(t)
		if err != nil {
			return err
		}
	}

	err = flagSet.Parse(os.Args[3:])
	if err != nil {
		return err
	}

	service := proto.NewBannersAPIProtobufClient(*serviceURLFlag, &http.Client{})

	request := &proto.UpdateBannerRequest{
		BannerId:  int32(*idFlag),
		ExpiresAt: timestamp,
		Message:   *messageFlag,
	}

	response, err := service.UpdateBanner(context.Background(), request)
	if err != nil {
		return err
	}

	json, err := json.Marshal(response.Banner)
	if err != nil {
		return err
	}

	fmt.Println(string(json))
	return nil
}

func bannerRemoveCommand(c *cmd.Command) error {
	flagSet := flag.NewFlagSet("banner remove", flag.ExitOnError)
	serviceURLFlag := newServiceURLFlag(flagSet)

	if len(os.Args) < 4 {
		return cmd.NewUsageError("usage: twirp-test banner remove <id>\n", flagSet)
	}

	// #nosec G109
	id, err := strconv.Atoi(os.Args[3])
	if err != nil {
		return err
	}

	err = flagSet.Parse(os.Args[3:])
	if err != nil {
		return err
	}

	service := proto.NewBannersAPIProtobufClient(*serviceURLFlag, &http.Client{})

	request := &proto.DeleteBannerRequest{
		BannerId: int32(id),
	}

	_, err = service.DeleteBanner(context.Background(), request)
	if err != nil {
		return err
	}

	fmt.Println("Banner removed")
	return nil
}

func bannerGetCommand(c *cmd.Command) error {
	flagSet := flag.NewFlagSet("banner get", flag.ExitOnError)
	serviceURLFlag := newServiceURLFlag(flagSet)

	if len(os.Args) < 4 {
		return cmd.NewUsageError("usage: twirp-test banner get <id>\n", flagSet)
	}

	// #nosec G109
	id, err := strconv.Atoi(os.Args[3])
	if err != nil {
		return err
	}

	err = flagSet.Parse(os.Args[3:])
	if err != nil {
		return err
	}

	service := proto.NewBannersAPIProtobufClient(*serviceURLFlag, &http.Client{})

	request := &proto.GetBannerRequest{
		BannerId: int32(id),
	}

	response, err := service.GetBanner(context.Background(), request)
	if err != nil {
		return err
	}

	if response.Banner != nil {
		json, err := json.Marshal(response.Banner)
		if err != nil {
			return err
		}

		fmt.Println(string(json))
	}

	return nil
}

func bannerListCommand(c *cmd.Command) error {
	flagSet := flag.NewFlagSet("banner list", flag.ExitOnError)
	idsFlag := flagSet.String("ids", "", "Optional list of banner ids to provide, separated by comma")
	serviceURLFlag := newServiceURLFlag(flagSet)

	if len(os.Args) < 3 {
		return cmd.NewUsageError("usage: twirp-test banner list [options]\n", flagSet)
	}

	err := flagSet.Parse(os.Args[3:])
	if err != nil {
		return err
	}

	var bannerIds []int32
	if idsFlag != nil && *idsFlag != "" {
		ids := strings.Split(*idsFlag, ",")
		for _, idstr := range ids {
			// #nosec G109
			id, err := strconv.Atoi(strings.TrimSpace(idstr))
			if err != nil {
				return fmt.Errorf("invalid format for -ids flag: '%s'", *idsFlag)
			}
			bannerIds = append(bannerIds, int32(id))
		}
	}

	service := proto.NewBannersAPIProtobufClient(*serviceURLFlag, &http.Client{})

	request := &proto.ListBannersRequest{
		BannerIds: bannerIds,
	}

	response, err := service.ListBanners(context.Background(), request)
	if err != nil {
		return err
	}

	if response.Banners != nil {
		json, err := json.Marshal(response.Banners)
		if err != nil {
			return err
		}

		fmt.Println(string(json))
	}

	return nil
}
