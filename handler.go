// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

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
	HandleHTTP(ctx context.Context, w ResponseWriter, r *http.Request) error
}

type HandlerFunc func(ctx context.Context, w ResponseWriter,
	r *http.Request) error

func (f HandlerFunc) HandleHTTP(ctx context.Context, w ResponseWriter,
	r *http.Request) error {
	return f(ctx, w, r)
}

// Base turns a webhelp.Handler into a http.Handler. You specify the root
// webhelp.Handler and an optional error recovery function.
type Base struct {
	// Root must be specified
	Root Handler

	// ErrHandler is optional. If unspecified, http.Error will be used with
	// the errhttp.GetErrorBody and errhttp.GetStatusCode values.
	ErrHandler func(w ResponseWriter, r *http.Request, err error)
}

func (b Base) handleError(w ResponseWriter, r *http.Request, err error) {
	if err == nil || w.WroteHeader() {
		return
	}
	if b.ErrHandler != nil {
		b.ErrHandler(w, r, err)
		return
	}
	http.Error(w, errhttp.GetErrorBody(err),
		errhttp.GetStatusCode(err, http.StatusInternalServerError))
}

func (b Base) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	nw, ok := w.(ResponseWriter)
	if !ok {
		nw = wrapResponseWriter(w)
	}
	if b.Root == nil {
		b.handleError(nw, r, ErrInternalServerError.New("No root handler installed"))
		return
	}
	ctx := context.Background()
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
	b.handleError(nw, r, b.Root.HandleHTTP(ctx, nw, r))
}

type loggingHandler struct {
	h Handler
}

// LoggingHandler takes a Handler and makes it log requests
func LoggingHandler(h Handler) Handler {
	return loggingHandler{h: h}
}

func (lh loggingHandler) HandleHTTP(ctx context.Context, w ResponseWriter,
	r *http.Request) error {
	logger.Noticef("%s %s", r.Method, r.RequestURI)
	err := lh.h.HandleHTTP(ctx, w, r)
	if err != nil {
		logger.Errorf("%s %s: %v", r.Method, r.RequestURI, err)
	}
	return err
}

func (lh loggingHandler) Routes(cb func(method, path string,
	annotations []string)) {
	Routes(lh.h, cb)
}

var _ RouteLister = loggingHandler{}

// StandardHandler takes an http.Handler and turns it into a webhelp.Handler
func StandardHandler(h http.Handler) Handler {
	return HandlerFunc(func(ctx context.Context, w ResponseWriter,
		r *http.Request) error {
		h.ServeHTTP(w, r)
		return nil
	})
}
