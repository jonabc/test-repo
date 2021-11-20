package rest

import (
	"context"
	"fmt"
	"net/http"

	"github.com/github/go-exceptions"
	"github.com/github/go-http-middleware/recovery"
	"github.com/github/go-log"
	"github.com/github/go-trace"
)

type key int

const (
	keyUserAgent key = iota
)

func logMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		log.Info(fmt.Sprintf("Request to %v", r.URL.Path))
		next.ServeHTTP(w, r)
		log.Info(fmt.Sprintf("End of request to %v", r.URL.Path))
	}

	return http.HandlerFunc(fn)
}

func traceMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")
		ctx := context.WithValue(context.Background(), keyUserAgent, ua)
		_, span := trace.ChildSpan(ctx, trace.OpFuncName(r.URL.Path))
		defer span.Finish()
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func getPanicMiddleware(reporter *exceptions.Reporter) func(http.Handler) http.Handler {
	rc := recovery.Recovery{
		Report: func(err error, req *http.Request) error {
			_ = reporter.Report(req.Context(), fmt.Errorf("panic: %+v", err), map[string]string{
				"method": req.Method,
				"url":    req.URL.String(),
			})
			return nil
		},
		Response: func(err error, rw http.ResponseWriter, r *http.Request) {
			rw.WriteHeader(500)
			fmt.Fprintf(rw, "insert sad robot here")
		},
	}

	return rc.Handler
}
