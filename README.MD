# apierr

This library is a utility for creating errors with the use of standard http status codes.
The built error is a `Problem` (from `schneider.vip/problem`).

The error handling on the api side will be simplified as follows:

```go
package main

import (
	"github.com/debyten/apierr"
	"net/http"
)

func exampleErr() error {
	return apierr.NotFound.Problem("entity not found")
}

func myHandler(w http.ResponseWriter, r *http.Request) {
	err := exampleErr()
	if err != nil {
        apierr.HandleISE(err, w)
		return
	}
}
```