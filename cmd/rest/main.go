package main

import (
	"fmt"
	"log"

	"github.com/jonabc/test-repo/gss/config"
	"github.com/jonabc/test-repo/gss/rest"
	"github.com/jonabc/test-repo/gss/server"
	"github.com/opentracing/opentracing-go"
	twtrace "github.com/twirp-ecosystem/twirp-opentracing"
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
	logger.Info("Initializing service...")

	reporter, err := cfg.NewExceptionReporter()
	if err != nil {
		return err
	}

	// initialize distributed tracing
	tracer := cfg.NewTracer()
	opentracing.SetGlobalTracer(tracer)

	// create new service
	handler := rest.NewHTTPHandler(reporter)
	// make sure we recover and propagate incoming tracing information
	// from HTTP requests
	handler = twtrace.WithTraceContext(handler, tracer)
	server := server.NewHTTPServer(handler, fmt.Sprintf(":%v", cfg.HTTPPort))

	// example log usage
	logger.Info("REST service initialized.")

	return server.Run()
}
