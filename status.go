package apierr

import (
	"fmt"
)

// Prefix used by all errors that must be presented to
// the clients.
const Prefix = "app.backend.errors"

// HttpStatus is a wrapper for http.StatusXXX proxying some apierr package methods.
//
// Example:
//
//	func (s InterfaceXXX) Execute(ctx context.Context, id string) error {
//		user, err = s.FindByID(ctx, id)
//		if err != nil {
//			return nil, NotFound.Err(err)
//		}
//		return user, nil
//	}
type HttpStatus int

// from http.StatusXXX

const (
	NotFound                      HttpStatus = 404
	BadRequest                    HttpStatus = 400
	InternalServerError           HttpStatus = 500
	Unauthorized                  HttpStatus = 401
	Forbidden                     HttpStatus = 403
	MethodNotAllowed              HttpStatus = 405
	NotAcceptable                 HttpStatus = 406
	RequestTimeout                HttpStatus = 408
	Conflict                      HttpStatus = 409
	Gone                          HttpStatus = 410
	LengthRequired                HttpStatus = 411
	PreconditionFailed            HttpStatus = 412
	RequestEntityTooLarge         HttpStatus = 413
	RequestURITooLong             HttpStatus = 414
	UnsupportedMediaType          HttpStatus = 415
	RequestedRangeNotSatisfiable  HttpStatus = 416
	ExpectationFailed             HttpStatus = 417
	Teapot                        HttpStatus = 418
	UnprocessableEntity           HttpStatus = 422
	Locked                        HttpStatus = 423
	FailedDependency              HttpStatus = 424
	UpgradeRequired               HttpStatus = 426
	PreconditionRequired          HttpStatus = 428
	TooManyRequests               HttpStatus = 429
	RequestHeaderFieldsTooLarge   HttpStatus = 431
	UnavailableForLegalReasons    HttpStatus = 451
	InternalServerErrorHttps      HttpStatus = 500
	NotImplemented                HttpStatus = 501
	BadGateway                    HttpStatus = 502
	ServiceUnavailable            HttpStatus = 503
	GatewayTimeout                HttpStatus = 504
	HttpVersionNotSupported       HttpStatus = 505
	VariantAlsoNegotiates         HttpStatus = 506
	InsufficientStorage           HttpStatus = 507
	LoopDetected                  HttpStatus = 508
	NotExtended                   HttpStatus = 510
	NetworkAuthenticationRequired HttpStatus = 511
)

// ErrTxt wraps apierr.FromText.
// The extra parameter can be empty.
func (h HttpStatus) ErrTxt(text string, extra ...any) error {
	return FromText(text, int(h), fmtArgs(extra)...)
}

// Err wraps apierr.New with the http status code and no extra.
func (h HttpStatus) Err(err error, extra ...any) error {
	return New(err, int(h), fmtArgs(extra)...)
}

// ClientErr wrap `text` using apierr.FromText and prepends Prefix to clientErr.
//
//	Example:
//	arg1 := 1
//	arg2 := "test"
//	japi.Conflict.ErrTxt("duplicate entry", "jarvis.backend.errors.myentity.myerrkey", arg1, arg2)
//
//	// without specify the prefix
//	japi.Conflict.ClientErr("duplicate entry", "myentity.myerrkey", arg1, arg2)
func (h HttpStatus) ClientErr(text string, clientErr string, extra ...any) error {
	key := fmtKey(clientErr)
	args := append([]any{key}, extra...)
	return h.ErrTxt(text, args...)
}

func fmtArgs(values []any) []string {
	formatted := make([]string, 0, len(values))
	for _, v := range values {
		formatted = append(formatted, fmt.Sprint(v))
	}
	return formatted
}

func fmtKey(clientErr string) string {
	return fmt.Sprintf("%s.%s", Prefix, clientErr)
}
