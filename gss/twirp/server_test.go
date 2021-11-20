package twirp

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jonabc/test-repo/gss"
	"github.com/jonabc/test-repo/gss/mocks"
	tu "github.com/jonabc/test-repo/gss/testutil"
	"github.com/jonabc/test-repo/gss/twirp/proto"
)

func TestAddBanner(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	mockBannerSvc := mocks.NewMockBannerService(ctrl)
	testServer := &Server{
		bannerSvc: mockBannerSvc,
	}

	mockBanner := &proto.Banner{
		BannerType: proto.Banner_Info,
		Message:    "The project will be undergoing regular maintenance on Saturday @ 9am PDT with no expected downtime",
	}

	mockBannerSvc.EXPECT().AddBanner(gomock.Any(), gss.BannerType(mockBanner.BannerType), nil, mockBanner.Message).Return(&gss.Banner{
		Message: mockBanner.Message,
	}, nil)

	req := &proto.AddBannerRequest{
		Banner: mockBanner,
	}

	res, err := testServer.AddBanner(ctx, req)
	tu.Ok(t, err)

	tu.Equals(t, "The project will be undergoing regular maintenance on Saturday @ 9am PDT with no expected downtime", res.Banner.Message)
}

func TestGetBanner(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	mockBannerSvc := mocks.NewMockBannerService(ctrl)
	testServer := &Server{
		bannerSvc: mockBannerSvc,
	}

	expiresAt := time.Now().Add(time.Hour * 24)

	mockBannerSvc.EXPECT().GetBanner(gomock.Any(), 1).Return(&gss.Banner{
		ID:        1,
		Type:      gss.BannerTypeWarning,
		ExpiresAt: &expiresAt,
		Message:   "This is a test of the emergency broadcast system",
	}, nil)

	req := &proto.GetBannerRequest{
		BannerId: 1,
	}

	res, err := testServer.GetBanner(ctx, req)
	tu.Ok(t, err)

	tu.NotEquals(t, nil, res.Banner)
	tu.Equals(t, "This is a test of the emergency broadcast system", res.Banner.Message)

	mockBannerSvc.EXPECT().GetBanner(gomock.Any(), 2).Return(nil, nil)

	req = &proto.GetBannerRequest{
		BannerId: 2,
	}

	res, err = testServer.GetBanner(ctx, req)
	tu.Ok(t, err)

	tu.Equals(t, (*proto.Banner)(nil), res.Banner)
}

func TestName(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	name := "Albus Dumbledore"
	mockWordSvc := mocks.NewMockWordService(ctrl)
	testServer := &Server{
		wordSvc: mockWordSvc,
	}

	req := &proto.NameRequest{
		Name: name,
	}

	res, err := testServer.HelloName(ctx, req)
	tu.Ok(t, err)

	expected := fmt.Sprintf("Hello, %v!", name)
	tu.Equals(t, expected, res.Message)
}

func TestReverseName(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	name := "Albus Dumbledore"
	reversedName := "erodelbmuD sublA"
	mockWordSvc := mocks.NewMockWordService(ctrl)
	mockWordSvc.EXPECT().ReverseWord(&gss.Word{
		Name: name,
	}).Return(&gss.Word{
		Name: reversedName,
	}, nil)
	testServer := &Server{
		wordSvc: mockWordSvc,
	}

	req := &proto.NameRequest{
		Name: name,
	}

	res, err := testServer.ReverseName(ctx, req)
	tu.Ok(t, err)
	tu.Equals(t, reversedName, res.Message)
}
