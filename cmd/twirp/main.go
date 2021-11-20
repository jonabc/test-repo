package main

import (
	"fmt"
	"log"

	"github.com/github/go-kvp"
	"github.com/jonabc/test-repo/gss/config"
	"github.com/jonabc/test-repo/gss/mem"
	"github.com/jonabc/test-repo/gss/mysql"
	"github.com/jonabc/test-repo/gss/mysql/sqlboiler"
	"github.com/jonabc/test-repo/gss/server"
	twirpserver "github.com/jonabc/test-repo/gss/twirp"
	"github.com/opentracing/opentracing-go"
)

func main() {
	if err := realMain(); err != nil {
		log.Fatalf("failed to run service: %v", err)
	}
}

func realMain() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := cfg.NewLogger()
	logger.Info("initializing service...")

	reporter, err := cfg.NewExceptionReporter()
	if err != nil {
		return err
	}

	// initialize statting
	statter, err := cfg.NewStatsClient()
	if err != nil {
		return err
	}
	logger.Info("starting statsd client.")
	statter.Run()
	defer statter.Stop()

	statter.Counter("gss.service.start", nil, 1)

	// initialize distributed tracing
	tracer := cfg.NewTracer()
	opentracing.SetGlobalTracer(tracer)

	// create new twirp service
	hooks, err := twirpserver.DefaultHooks(logger, reporter, statter, tracer)
	if err != nil {
		return err
	}

	db, err := mysql.OpenDB(cfg)
	if err != nil {
		return err
	}

	// Choose from the following banner service implementation
	// Vanilla SQL or SqlBoiler
	bannerSvc := sqlboiler.NewBannerService(db)
	// bannerSvc := sqlx.NewBannerService(db)
	// bannerSvc := vanilla.NewBannerService(db)

	wordSvc := mem.NewWordService()

	twirpServer, err := twirpserver.NewTwirpServer(hooks, wordSvc, bannerSvc)
	if err != nil {
		return err
	}

	server := server.NewHTTPServer(twirpServer, fmt.Sprintf(":%v", cfg.HTTPPort))

	// example log usage: use key-values instead of `printf` string interpolation,
	// to facilitate searching of logs with key-value arguments.
	logger.Info("twirp service initialized in environment", kvp.String("environment", cfg.Environment))

	return server.Run()
}
