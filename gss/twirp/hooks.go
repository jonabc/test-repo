package twirp

import (
	"context"

	"github.com/github/go-exceptions"
	"github.com/github/go-kvp"
	"github.com/github/go-log"
	"github.com/github/go-stats"
	twhooks "github.com/github/go-twirp/server/hooks"
	twlog "github.com/github/go-twirp/server/hooks/log"
	twstats "github.com/github/go-twirp/server/hooks/stats"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	twtrace "github.com/twirp-ecosystem/twirp-opentracing"
	"github.com/twitchtv/twirp"
)

// DefaultHooks returns a set of recommended hooks.
func DefaultHooks(logger *log.Logger, reporter *exceptions.Reporter, statter stats.Client, tracer opentracing.Tracer) (*twirp.ServerHooks, error) {
	hooks := twirp.ChainHooks(
		twhooks.TimingHooks(),
		twhooks.StoreTwirpErrorHooks(),
		twlog.DefaultHooks(logger),
		twtrace.NewOpenTracingHooks(tracer),
		twstats.DefaultHooks(statter),
		errorReporterHook(logger, reporter),
		statsReporterHook(statter),
	)

	return hooks, nil
}

func statsReporterHook(statter stats.Client) *twirp.ServerHooks {
	return &twirp.ServerHooks{
		RequestReceived: func(ctx context.Context) (context.Context, error) {
			tags := stats.Tags{}
			if m, ok := twirp.MethodName(ctx); ok {
				tags["twirp_method"] = m
			}
			if m, ok := twirp.PackageName(ctx); ok {
				tags["twirp_package_name"] = m
			}
			if m, ok := twirp.ServiceName(ctx); ok {
				tags["twirp_service_name"] = m
			}

			statter.Counter("request.received", tags, 1)

			return ctx, nil
		},
	}
}

// errorReporterHook reports an errors with the given reporter.
func errorReporterHook(logger *log.Logger, reporter *exceptions.Reporter) *twirp.ServerHooks {
	return &twirp.ServerHooks{
		Error: func(ctx context.Context, twerr twirp.Error) context.Context {
			payload := map[string]string{}

			if m, ok := twirp.MethodName(ctx); ok {
				payload["twirp_method"] = m
			}
			if m, ok := twirp.PackageName(ctx); ok {
				payload["twirp_package_name"] = m
			}
			if m, ok := twirp.ServiceName(ctx); ok {
				payload["twirp_service_name"] = m
			}
			if m, ok := twirp.StatusCode(ctx); ok {
				payload["http_status"] = m
			}

			// twirp wraps errors for pkg/errors interoperability whenever we
			// use "twirp.InternalErrorWith()". This allows us to return the
			// underlying error.
			err := errors.Cause(twerr)

			if err := reporter.Report(ctx, err, payload); err != nil {
				logger.Error("ERROR: Failed to report error.", kvp.Err(err))
			}

			return ctx
		},
	}
}
