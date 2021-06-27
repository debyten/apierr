package apierr

import (
	"errors"
	"net/http"
	"strings"
)

// ErrHeader is the header key written on response
const ErrHeader = "X-App-Error"

// DBNotFoundHandler is the checker function used from HandleISE for
// DB record not found errors. DefaultDBNotFoundHandler should be overridden
type DBNotFoundHandler func(err error) bool

// DefaultDBNotFoundHandler override this function to correctly handle
// Database errors
var DefaultDBNotFoundHandler DBNotFoundHandler = func(_ error) bool {
	return false
}

// Handle check if error is of type APIError (using errors.As and errors.Unwrap functions)
// and writes the following information on ResponseWriter:
//
//  - APIErr.StatusCode(): the http status code
//  - http.StatusText(APIErr.StatusCode()): the http status text from status code
//  - ErrHeader header if APIErr.Extra() is present (see APIErr for more details)
//
// When the status code is not a client or server error (e.g. 204 no content) it will be handled without the use of http.Error.
// The http.ResponseWriter WriteHeader function will be invoked.
//
// returns true if err is of type APIErr.
func Handle(err error, w http.ResponseWriter) bool {
	ae, ok := extractAPIErr(err)
	if !ok {
		return false
	}
	if ae.extras {
		w.Header().Add(ErrHeader, ae.mergeExtra())
	}
	for k, values := range ae.customHeaders {
		w.Header().Add(k, strings.Join(values, ","))
	}
	statusCode := ae.StatusCode()
	// handle with http.Error only if is a client or server error
	if statusCode >= 400 && statusCode < 600 {
		http.Error(w, http.StatusText(statusCode), statusCode)
		return true
	}
	w.WriteHeader(statusCode)
	return true
}

// HandleISE executes Handle().
// If Handle() returns false executes DefaultDBNotFoundHandler that handles common db "not found" errors.
//
// If the error is unknown (not an APIErr nor a DBNotFoundErr) it will reply with Internal Server Error.
func HandleISE(err error, w http.ResponseWriter, r *http.Request) {
	for _, decorator := range decorators {
		decorator(w, r)
	}
	if Handle(err, w) {
		return
	}
	if DefaultDBNotFoundHandler(err) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func extractAPIErr(err error) (*APIErr, bool) {
	var ae *APIErr
	for err != nil {
		if errors.As(err, &ae) {
			return ae, true
		}
		err = errors.Unwrap(err)
	}
	return nil, false
}