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
//
// Deprecated: use AddHandler
var DefaultDBNotFoundHandler DBNotFoundHandler = func(_ error) bool {
	return false
}

// An ErrHandler is a function which takes an err as input and eventually returns a problem.Problem.
//
// Custom err matching logic can be implemented in the handler function. The implemented function
// must return nil if the error doesn't match the implemented logic, otherwise it should explicitly return
// a problem.Problem.
type ErrHandler func(err error) *problem.Problem

var handlers []ErrHandler

// AddHandler to the known apierr handlers.
func AddHandler(handler func(err error) *problem.Problem) {
	handlers = append(handlers, handler)
}

// Handle processes errors as problem.Problem instances. It recursively unwraps the error
// until either a problem.Problem is found or the error is nil.
//
// If a problem.Problem is found, it is written to the response writer.
// If not, the error is passed through custom error handlers. If any handler matches,
// the resulting problem is written to the response writer and the function returns true.
// Otherwise, http.StatusInternalServerError is written to the response and the function returns false.
func Handle(err error, w http.ResponseWriter) bool {
	ae := extractProblem(err)
	if ae != nil {
		_, _ = ae.WriteTo(w)
		return true
	}
	for _, handler := range handlers {
		if p := handler(err); p != nil {
			_, _ = p.WriteTo(w)
			return true
		}
	}
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	return false
}

// HandleISE executes Handle.
// When Handle return false then executes DefaultDBNotFoundHandler; the last one handles common db "not found" errors.
//
// If the error is unknown (not a Problem nor a DBNotFoundErr) it will reply with Internal Server Error.
//
// Deprecated: use Handle
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
