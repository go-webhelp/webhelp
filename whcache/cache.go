// Copyright (C) 2017 JT Olds
// See LICENSE for copying information

// Package whcache provides a mechanism for per-request computation caching
//
// Sometimes you have a helper method that performs a computation or interacts
// with a datastore or remote resource, and that helper method gets called
// repeatedly. With simple context values, since the helper method is the most
// descendent frame in the stack, there isn't a context specific place
// (besides maybe the session, which would be a bad choice) to cache context
// specific values. With this cache helper, there now is.
//
// The cache must be registered up the handler chain with Register, and then
// helper methods can use Set/Get/Remove to interact with the cache (if
// available). If no cache was registered, Set/Get/Remove will still work, but
// Get will never return a value.
package whcache // import "gopkg.in/webhelp.v1/whcache"

import (
	"net/http"

	"golang.org/x/net/context"
	"gopkg.in/webhelp.v1"
	"gopkg.in/webhelp.v1/whcompat"
	"gopkg.in/webhelp.v1/whroute"
)

var (
	cacheKey = webhelp.GenSym()
)

type reqCache map[interface{}]interface{}

// Register installs a cache in the handler chain.
func Register(h http.Handler) http.Handler {
	return whroute.HandlerFunc(h, func(w http.ResponseWriter, r *http.Request) {
		ctx := whcompat.Context(r)
		if _, ok := ctx.Value(cacheKey).(reqCache); ok {
			h.ServeHTTP(w, r)
			return
		}
		h.ServeHTTP(w, whcompat.WithContext(r,
			context.WithValue(ctx, cacheKey, reqCache{})))
	})
}

// Set stores the key/val pair in the context specific cache, if possible.
func Set(ctx context.Context, key, val interface{}) {
	cache, ok := ctx.Value(cacheKey).(reqCache)
	if !ok {
		return
	}
	cache[key] = val
}

// Remove removes any values stored with key from the context specific cache,
// if possible.
func Remove(ctx context.Context, key interface{}) {
	cache, ok := ctx.Value(cacheKey).(reqCache)
	if !ok {
		return
	}
	delete(cache, key)
}

// Get returns previously stored key/value pairs from the context specific
// cache if one is registered and the value is found, and returns nil
// otherwise.
func Get(ctx context.Context, key interface{}) interface{} {
	cache, ok := ctx.Value(cacheKey).(reqCache)
	if !ok {
		return nil
	}
	val, _ := cache[key]
	return val
}
