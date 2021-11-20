package twirp

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/github/go-trace"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/jonabc/test-repo/gss"
	"github.com/jonabc/test-repo/gss/twirp/proto"
	"github.com/twitchtv/twirp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server holds the methods for our Twirp Server as well as a few dependencies.
type Server struct {
	wordSvc   gss.WordService
	bannerSvc gss.BannerService
}

// NewTwirpServer creates a new Twirp server.
func NewTwirpServer(hooks *twirp.ServerHooks, wordSvc gss.WordService, bannerSvc gss.BannerService) (http.Handler, error) {
	server := &Server{
		wordSvc:   wordSvc,
		bannerSvc: bannerSvc,
	}

	server1 := proto.NewHelloWorldAPIServer(server, hooks)
	server2 := proto.NewBannersAPIServer(server, hooks)

	mux := http.NewServeMux()
	mux.Handle(server1.PathPrefix(), server1)
	mux.Handle(server2.PathPrefix(), server2)
	mux.HandleFunc("/_ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	})

	return mux, nil
}

// HelloName will echo whatever name it is given in a canned message.
func (s *Server) HelloName(ctx context.Context, req *proto.NameRequest) (*proto.NameResponse, error) {
	_, span := trace.ChildSpan(ctx)
	defer span.Finish()

	message := fmt.Sprintf("Hello, %v!", req.Name)
	return &proto.NameResponse{
		Message: message,
	}, nil
}

// ReverseName will return the name but reversed.
func (s *Server) ReverseName(ctx context.Context, req *proto.NameRequest) (*proto.NameResponse, error) {
	_, span := trace.ChildSpan(ctx)
	defer span.Finish()

	msg, err := s.wordSvc.ReverseWord(&gss.Word{
		Name: req.Name,
	})

	if err != nil {
		return nil, err
	}

	return &proto.NameResponse{
		Message: msg.Name,
	}, nil
}

// AddBanner persists that using the BannerService.
func (s *Server) AddBanner(ctx context.Context, req *proto.AddBannerRequest) (*proto.AddBannerResponse, error) {
	ctx, span := trace.ChildSpan(ctx)
	defer span.Finish()

	bt := gss.BannerType(req.GetBanner().GetBannerType())
	msg := req.GetBanner().GetMessage()
	if msg == "" {
		return nil, span.WithError(twirp.RequiredArgumentError("banner.message"))
	}

	var expiresAt *time.Time
	if req.GetBanner().GetExpiresAt() != nil {
		time := req.GetBanner().GetExpiresAt()
		if err := time.CheckValid(); err != nil {
			return nil, span.WithError(err)
		}
		eat := req.GetBanner().GetExpiresAt().AsTime()
		expiresAt = &eat
	}

	banner, err := s.bannerSvc.AddBanner(ctx, bt, expiresAt, msg)
	if err != nil {
		return nil, span.WithError(err)
	}

	bannerProto, err := toProtoBanner(banner)
	if err != nil {
		return nil, span.WithError(err)
	}

	return &proto.AddBannerResponse{Banner: bannerProto}, nil
}

// DeleteBanner deletes the banner identified by ID with the BannerService.
func (s *Server) DeleteBanner(ctx context.Context, req *proto.DeleteBannerRequest) (*proto.DeleteBannerResponse, error) {
	ctx, span := trace.ChildSpan(ctx)
	defer span.Finish()

	err := s.bannerSvc.DeleteBanner(ctx, int(req.GetBannerId()))
	if err != nil {
		return nil, span.WithError(err)
	}

	return &proto.DeleteBannerResponse{}, nil
}

// UpdateBanner updates the requested banner with the BannerService.
func (s *Server) UpdateBanner(ctx context.Context, req *proto.UpdateBannerRequest) (*proto.UpdateBannerResponse, error) {
	ctx, span := trace.ChildSpan(ctx)
	defer span.Finish()

	var expiresAt *time.Time
	if req.GetExpiresAt() != nil {
		time := req.GetExpiresAt()
		if err := time.CheckValid(); err != nil {
			return nil, span.WithError(err)
		}
		eat := time.AsTime()
		expiresAt = &eat
	}

	banner := &gss.Banner{
		ID:        int(req.GetBannerId()),
		ExpiresAt: expiresAt,
		Message:   req.GetMessage(),
	}

	updatedBanner, err := s.bannerSvc.UpdateBanner(ctx, banner)
	if err != nil {
		return nil, span.WithError(err)
	}

	updatedBannerProto, err := toProtoBanner(updatedBanner)
	if err != nil {
		return nil, span.WithError(err)
	}

	return &proto.UpdateBannerResponse{Banner: updatedBannerProto}, nil
}

// GetBanner accepts an ID and returns the requested banner from the BannerService.
func (s *Server) GetBanner(ctx context.Context, req *proto.GetBannerRequest) (*proto.GetBannerResponse, error) {
	ctx, span := trace.ChildSpan(ctx)
	defer span.Finish()

	banner, err := s.bannerSvc.GetBanner(ctx, int(req.GetBannerId()))
	if err != nil {
		return nil, span.WithError(err)
	}

	bannerProto, err := toProtoBanner(banner)
	if err != nil {
		return nil, span.WithError(err)
	}

	return &proto.GetBannerResponse{Banner: bannerProto}, nil
}

// ListBanners will return all banners in the BannerService.
func (s *Server) ListBanners(ctx context.Context, req *proto.ListBannersRequest) (*proto.ListBannersResponse, error) {
	ctx, span := trace.ChildSpan(ctx)
	defer span.Finish()

	var ids []int
	if req.BannerIds != nil {
		ids = make([]int, len(req.BannerIds))
		for i, v := range req.BannerIds {
			ids[i] = int(v)
		}
	}

	banners, err := s.bannerSvc.ListBanners(ctx, ids)
	if err != nil {
		return nil, span.WithError(err)
	}

	bannersProto := make([]*proto.Banner, len(banners))
	for i, v := range banners {
		bannersProto[i], err = toProtoBanner(v)
		if err != nil {
			return nil, span.WithError(err)
		}
	}

	return &proto.ListBannersResponse{Banners: bannersProto}, nil
}

func toProtoBanner(b *gss.Banner) (*proto.Banner, error) {
	if b == nil {
		return nil, nil
	}

	var (
		expiresAt *timestamp.Timestamp
		createdAt *timestamp.Timestamp
		err       error
	)

	if b.ExpiresAt != nil {
		expiresAt = timestamppb.New(*b.ExpiresAt)
		if err != nil {
			return nil, err
		}
	}

	if b.CreatedAt != nil {
		createdAt = timestamppb.New(*b.CreatedAt)
		if err != nil {
			return nil, err
		}
	}

	return &proto.Banner{
		BannerId:   int32(b.ID),
		BannerType: proto.Banner_BannerType(b.Type),
		ExpiresAt:  expiresAt,
		CreatedAt:  createdAt,
		RepoId:     int32(b.RepositoryID),
		CreatorId:  int32(b.CreatorID),
		Message:    b.Message,
	}, nil
}
