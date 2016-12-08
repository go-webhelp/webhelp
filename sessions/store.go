// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

// package sessions is a lightweight session storage mechanism for the webhelp
// package. Attempting to be a combination of minimal and useful. Implementing
// the Store interface is all one must do to provide a different session
// storage mechanism.
package sessions

import (
	"net/http"

	"github.com/jtolds/webhelp"
	"github.com/spacemonkeygo/errors"
	"golang.org/x/net/context"
)

type ctxKey int

var (
	reqCtxKey ctxKey = 1
)

type SessionData struct {
	New    bool
	Values map[interface{}]interface{}
}

type Session struct {
	SessionData

	store     Store
	namespace string
}

type Store interface {
	Load(r *http.Request, namespace string) (SessionData, error)
	Save(w http.ResponseWriter, namespace string, s SessionData) error
	Clear(w http.ResponseWriter, namespace string) error
}

type reqCtx struct {
	s     Store
	r     *http.Request
	cache map[string]*Session
}

// HandlerWithStore wraps a webhelp.Handler such that Load works with contexts
// provided in that Handler.
func HandlerWithStore(s Store, h http.Handler) http.Handler {
	return webhelp.RouteHandlerFunc(h,
		func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, webhelp.WithContext(r, context.WithValue(
				webhelp.Context(r), reqCtxKey, &reqCtx{s: s, r: r})))
		})
}

// Load will return the current session, creating one if necessary. This will
// fail if a store wasn't installed with HandlerWithStore somewhere up the
// call chain.
func Load(ctx context.Context, namespace string) (*Session, error) {
	rc, ok := ctx.Value(reqCtxKey).(*reqCtx)
	if !ok {
		return nil, errors.ProgrammerError.New(
			"no session store handler wrapper installed")
	}
	if rc.cache != nil {
		if session, exists := rc.cache[namespace]; exists {
			return session, nil
		}
	}
	sessiondata, err := rc.s.Load(rc.r, namespace)
	if err != nil {
		return nil, err
	}
	session := &Session{
		SessionData: sessiondata,
		store:       rc.s,
		namespace:   namespace}
	if rc.cache == nil {
		rc.cache = map[string]*Session{namespace: session}
	} else {
		rc.cache[namespace] = session
	}
	return session, nil
}

// Save saves the session using the appropriate mechanism.
func (s *Session) Save(w http.ResponseWriter) error {
	err := s.store.Save(w, s.namespace, s.SessionData)
	if err == nil {
		s.SessionData.New = false
	}
	return err
}

// Clear clears the session using the appropriate mechanism.
func (s *Session) Clear(w http.ResponseWriter) error {
	// clear out the cache
	for name := range s.Values {
		delete(s.Values, name)
	}
	return s.store.Clear(w, s.namespace)
}
