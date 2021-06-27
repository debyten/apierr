package apierr

import "net/http"

// Decorator should be used to add custom details on http.ResponseWriter and will be executed only for Unhandled errors
// on HandleISE.
//  Example:
//    myErrDecorator := func(w http.ResponseWriter, r *http.Request) {
//      v, ok := r.Context().Value(myKeyVal)
//        if !ok {
//         return
//       }
//      w.Header().Add("X-MY-ERR-HEADER", v)
//    }
type Decorator func(w http.ResponseWriter, r *http.Request)

var decorators = make([]Decorator, 0)

// AddDecorator adds a custom decorator
func AddDecorator(decorator Decorator) {
	decorators = append(decorators, decorator)
}
