package webhelp

import (
	"net/http"

	"github.com/spacemonkeygo/errors/errhttp"
	"github.com/spacemonkeygo/spacelog"
	"golang.org/x/net/context"
)

var (
	logger = spacelog.GetLogger()
)

// Handler is extended to include a Context, a ResponseWriter with some useful
// additional methods, and return an error. Errors can be tagged with status
// codes and messages using github.com/spacemonkeygo/errors/errhttp.
type Handler interface {
	ServeHTTP(ctx context.Context, w ResponseWriter, r *http.Request) error
}

type HandlerFunc func(ctx context.Context, w ResponseWriter,
	r *http.Request) error

func (f HandlerFunc) ServeHTTP(ctx context.Context, w ResponseWriter,
	r *http.Request) error {
	return f(ctx, w, r)
}

func toStandardHandler(h Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		nw, ok := w.(ResponseWriter)
		if !ok {
			nw = wrapResponseWriter(w)
		}
		if cnw, ok := nw.(http.CloseNotifier); ok {
			doneChan := make(chan bool)
			defer close(doneChan)
			closeChan := cnw.CloseNotify()
			var cancelFunc func()
			ctx, cancelFunc = context.WithCancel(ctx)
			go func() {
				select {
				case <-doneChan:
					cancelFunc()
				case <-closeChan:
					cancelFunc()
				}
			}()
		}
		err := h.ServeHTTP(ctx, nw, r)
		if err != nil && !nw.WroteHeader() {
			http.Error(nw, errhttp.GetErrorBody(err),
				errhttp.GetStatusCode(err, http.StatusInternalServerError))
		}
	})
}

// LoggingHandler takes a Handler and makes it log requests
func LoggingHandler(h Handler) Handler {
	return HandlerFunc(func(ctx context.Context, w ResponseWriter,
		r *http.Request) error {
		logger.Noticef("%s %s", r.Method, r.RequestURI)
		err := h.ServeHTTP(ctx, w, r)
		if err != nil {
			logger.Errorf("%s %s: %v", r.Method, r.RequestURI, err)
		}
		return err
	})
}

// StandardHandler takes an http.Handler and turns it into a webhelp.Handler
func StandardHandler(h http.Handler) Handler {
	return HandlerFunc(func(ctx context.Context, w ResponseWriter,
		r *http.Request) error {
		h.ServeHTTP(w, r)
		return nil
	})
}
