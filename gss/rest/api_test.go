package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	tu "github.com/jonabc/test-repo/gss/testutil"
)

// TestHelloHandler tests the helloHandler from the api package
// testing code stolen from Go maintainer Matt Silverlock at https://blog.questionable.services/article/testing-http-handlers-go/
func TestHelloHandler(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/hello", nil)
	tu.Ok(t, err)

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(helloHandler)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		tu.Equals(t, http.StatusOK, status)
	}

	// Check the response body is what we expect.
	expected := `Hello world!`
	if rr.Body.String() != expected {
		tu.Equals(t, expected, rr.Body.String())
	}
}
