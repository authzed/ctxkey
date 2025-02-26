// This is a simple example of using ctxkey with http middleware.
// It demonstrates all three types of keys: normal, default, and boxed
// and the ways they can be used to pass values through a chain of middleware.

package main

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/authzed/ctxkey"
)

type User struct {
	ID   int
	Name string
}

const name = "Alice"

var (
	// ctxAuthorizedUser is a context key that stores a typed User representing the authorized user
	ctxAuthorizedUser = ctxkey.New[User]()

	// ctxLogger is a context key that stores a slog.Logger, but will return a default logger if unset
	ctxLogger = ctxkey.NewWithDefault(slog.Default())

	// ctxBytesWritten is a context key that stores the number of bytes written by a handler
	// the NewBoxedWithDefault type is used when "decorating". In this case, a handler will
	// fill in the value lower down the chain to be read by a wrapping middleware.
	ctxBytesWritten = ctxkey.NewBoxedWithDefault(0)
)

// middleware fills in the user key after authorization
func authorizeUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxLogger.MustNonEmptyValue(r.Context()).Info("authorizing user", "name", name)

		user := &User{ID: 1, Name: name}
		r = r.WithContext(ctxAuthorizedUser.Set(r.Context(), *user))
		next.ServeHTTP(w, r)
	})
}

// bytesWrittenLoggingMiddleware logs how many bytes were written by a handler
func bytesWrittenLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// carve out a spot in the context for the value to be written
		r = r.WithContext(ctxBytesWritten.SetBox(r.Context()))

		// run the wrapped handler
		next.ServeHTTP(w, r)

		// extract the value from the context and log it
		bytesWritten := ctxBytesWritten.Value(r.Context())
		ctxLogger.MustNonEmptyValue(r.Context()).Info("wrote response", "bytes", bytesWritten)
	})
}

var helloHandler http.HandlerFunc = func(w http.ResponseWriter, req *http.Request) {
	// get user from context, will panic if missing
	user := ctxAuthorizedUser.MustValue(req.Context())

	written, err := fmt.Fprintf(w, "hello %s\n", user.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// fill in the "box" that is already present in the ctx
	ctxBytesWritten.Set(req.Context(), written)
}

func main() {
	mux := http.NewServeMux()

	// set up routes
	mux.Handle("/hello", helloHandler)

	// install middleware
	handler := bytesWrittenLoggingMiddleware(mux)
	handler = authorizeUserMiddleware(handler)

	// serve
	if err := http.ListenAndServe(":8090", handler); err != nil {
		panic(err)
	}
}
