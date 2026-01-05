package chiErrorHandler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path"
	"runtime"

	"github.com/go-chi/chi/v5/middleware"
)

type httpError struct {
	err   error
	stack string
}

type httpErrors struct {
	errors []httpError
}

// AttachError attaches an error to the request context for logging purposes.
// It captures the caller's file and line number.
func AttachError(r *http.Request, err error) {
	_, file, line, ok := runtime.Caller(1)
	if ok {
		file = fmt.Sprintf("%s:%d", path.Base(file), line)
	} else {
		file = "unable to determine caller info"
	}
	r.Context().Value("errors").(*httpErrors).attach(httpError{
		err:   err,
		stack: file,
	})
}

func (e *httpErrors) attach(err httpError) {
	e.errors = append(e.errors, err)
}

// ErrorLogger is a middleware that logs errors attached to the request context.
func ErrorLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(context.WithValue(r.Context(), "errors", &httpErrors{}))

		next.ServeHTTP(w, r)
		ctx := r.Context()
		reqID := middleware.GetReqID(ctx)
		if errs, ok := ctx.Value("errors").(*httpErrors); ok {
			for _, err := range errs.errors {
				log.Printf("error handling request %s: %s %s\n", reqID, err.err.Error(), err.stack)
			}
		}
	})
}
