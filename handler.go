package apierr

import (
	"errors"
	"net/http"
	"schneider.vip/problem"
)

// DBNotFoundHandler is the checker function used from HandleISE for
// DB record not found errors. DefaultDBNotFoundHandler should be overridden
type DBNotFoundHandler func(err error) bool

// DefaultDBNotFoundHandler override this function to correctly handle
// Database errors
var DefaultDBNotFoundHandler DBNotFoundHandler = func(_ error) bool {
	return false
}

// Handle err as a problem.Problem. The error is unwrapped recursively until is nil.
// If a problem.Problem is not found, then return false, otherwise writes the response
// and return true.
func Handle(err error, w http.ResponseWriter) bool {
	ae := extractProblem(err)
	if ae == nil {
		return false
	}
	_, _ = ae.WriteTo(w)
	return true
}

// HandleISE executes Handle.
// When Handle return false then executes DefaultDBNotFoundHandler; the last one handles common db "not found" errors.
//
// If the error is unknown (not a Problem nor a DBNotFoundErr) it will reply with Internal Server Error.
func HandleISE(err error, w http.ResponseWriter) {
	if Handle(err, w) {
		return
	}
	if DefaultDBNotFoundHandler(err) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func extractProblem(err error) *problem.Problem {
	var ae *problem.Problem
	for err != nil {
		if errors.As(err, &ae) {
			return ae
		}
		err = errors.Unwrap(err)
	}
	return nil
}
