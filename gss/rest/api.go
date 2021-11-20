package rest

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/github/go-exceptions"
	"github.com/github/go-kvp"
	"github.com/github/go-log"
	"github.com/github/go-trace"
	"github.com/justinas/alice"
)

// NewHTTPHandler compiles a mux of all the handlers and returns it to be passed to the server.
func NewHTTPHandler(reporter *exceptions.Reporter) http.Handler {
	// alice allows us to chain a series of handlers together
	// in this case, we have a log middleware, a trace middleware, and a panic middleware
	// the log middleware logs the request and then calls the next handler in the chain (the trace middleware)
	// the trace middleware sends a span of the duration of the request to LightStep and then calls the next handler in the chain (the panic middleware)
	// the panic middleware sets up a panic handler and calls the next handler in the chain, gracefully recording any panic
	mw := alice.New(logMiddleware, traceMiddleware, getPanicMiddleware(reporter))

	base := mw.Then(http.HandlerFunc(baseHandler))
	hello := mw.Then(http.HandlerFunc(helloHandler))
	echo := mw.Then(http.HandlerFunc(echoHandler))
	panic := mw.Then(http.HandlerFunc(panicHandler))
	longOp := mw.Then(http.HandlerFunc(longOpHandler))

	mux := http.NewServeMux()
	mux.Handle("/", base)
	mux.Handle("/hello", hello)
	mux.Handle("/echo", echo)
	mux.Handle("/panic", panic)
	mux.Handle("/long-op", longOp)
	return mux
}

func baseHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Example response.")
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world!")
}

func longOpHandler(w http.ResponseWriter, r *http.Request) {
	// make a channel to wait for upcoming worker goroutine
	start := time.Now()
	finished := make(chan bool)

	// start a dumb goroutine
	go longOpWorker(r.Context(), finished)

	// wait for goroutine to finish and return execution time
	<-finished
	fmt.Fprintf(w, "Long operation executed in %s", time.Since(start))
}

func longOpWorker(ctx context.Context, finished chan bool) {
	// trace worker execution
	_, span := trace.ChildSpan(ctx, trace.OpFuncName("long-op-worker"))
	defer span.Finish()

	// just wait random amount of time to simulate a long operation
	// mostly used for tracing reporting, to see a bit of variation in
	// request response times.
	dur, _ := rand.Int(rand.Reader, big.NewInt(100))
	time.Sleep(time.Duration(dur.Int64()) * time.Millisecond)
	finished <- true
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "error: /echo only accepts POST requests.", http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var t interface{}
	decodeErr := decoder.Decode(&t)
	if decodeErr != nil {
		http.Error(w, "error: failed to parse JSON from request body.", http.StatusBadRequest)
		return
	}

	encodeErr := json.NewEncoder(w).Encode(t)
	if encodeErr != nil {
		log.Error("error encoding JSON to return.", kvp.Err(encodeErr))
		return
	}
}

func panicHandler(w http.ResponseWriter, r *http.Request) {
	panic("/panic endpoint called. AHHHHHHHH")
}
