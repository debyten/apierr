package apierr

import (
	"errors"
	"strings"
)

// New returns an APIErr with the defined err and status code. If extra parameters are present
// these parameters will be put in the response headers with the key ErrHeader.
//  Example:
//  aErr := apierr.New(err, http.StatusBadRequest, "malformed id", "BAD_INPUT_ID")
//  // the default ErrHeader is X-App-Error, so the resulting http response
//  // for the following request PUT /users/BAD_INPUT_ID is:
//
//  HTTP/1.1 400 Bad request
//  X-App-Error: malformed id,BAD_INPUT_ID
//  ... more headers ...
func New(err error, code int, extra ...string) *APIErr {
	return &APIErr{
		err: err,
		extra: extra,
		code: code,
		extras: len(extra) > 0,
		customHeaders: make(map[string][]string),
	}
}

// FromText creates a new error with errors.New(errText) and invokes New.
func FromText(errText string, code int, extra ...string) *APIErr {
	err := errors.New(errText)
	return New(err, code, extra...)
}

// APIErr is an error with StatusCode(), which should be an http status code
//
// Extra() could be useful to return information about the error on the response.
//  Example of built header:
//    X-App-Error: "my.err.to.client,VALUE1,VALUE2..."
//
// This type also, can be used from services.
//  Example of usage:
//    var errInvalidAge = fmt.Errorf("invalid age")
//    var errSameAge = fmt.Errorf("same age")
//    func (svc *UserServiceImpl) UpdateUserAge(id, age int) error {
//      user, err := svc.repository.FindByID(id)
//      // if err != nil ...
//      if user.age == age {
//        return apierr.New(errSameAge, http.StatusConflict)
//      }
//      // and so on...
//    }
//
//  From resource:
//    func (api *UserApiImpl) HandleChangeAge(w http.ResponseWriter, r *http.Request) {
//      // retrieve id, age...
//      if err := api.userService.UpdateUserAge(id, age); err != nil {
//        apierr.HandleISE(err, w)
//        return
//      }
//    }
type APIErr struct {
	code          int
	err           error
	extra         []string
	customHeaders map[string][]string
	extras        bool
}

// CustomHeader adds a custom header, it works in append mode:
//  Example:
//    New(err, 400).
//     CustomHeader("X-My-Custom-Header", "value1").
//     CustomHeader("X-My-Custom-Header", "value2")
//  Result:
//  X-My-Custom-Header: "value1, value2"
func (a *APIErr) CustomHeader(k string, v string) *APIErr {
	curr, ok := a.customHeaders[k]
	if !ok {
		curr = make([]string, 0)
	}
	curr = append(curr, v)
	a.customHeaders[k] = curr
	return a
}

func (a *APIErr) Error() string {
	if a.err == nil {
		return ""
	}
	return a.err.Error()
}

// StatusCode returns the http status code. This signature is the same of StatusCoder interface
// in the go-kit lib (https://github.com/go-kit/kit/blob/0d7a3880d126d0a090d817367d189c95a455c0ec/transport/http/server.go#L216)
func (a *APIErr) StatusCode() int {
	return a.code
}

func (a *APIErr) mergeExtra() string {
	return strings.Join(a.extra, ",")
}