// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package whmon

import (
	"math/rand"
	"net/http"
	"sync/atomic"
	"time"

	"gopkg.in/webhelp.v1"
	"gopkg.in/webhelp.v1/whcompat"
	"gopkg.in/webhelp.v1/whroute"
	"golang.org/x/net/context"
)

var (
	RequestId = webhelp.GenSym()

	idCounter uint64
	inc       uint64
)

// RequestIds generates a new request id for the request if one does not
// already exist under the Request Context Key RequestId.
// The RequestId can be retrieved using:
//
//   rid := whcompat.Context(req).Value(whmon.RequestId).(int64)
//
func RequestIds(h http.Handler) http.Handler {
	return addKey(h, RequestId, func(r *http.Request) interface{} {
		rid, ok := whcompat.Context(r).Value(RequestId).(int64)
		if !ok {
			rid = newId()
		}
		return rid
	})
}

func addKey(h http.Handler, key interface{},
	val func(r *http.Request) interface{}) http.Handler {
	return whroute.HandlerFunc(h, func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, whcompat.WithContext(r,
			context.WithValue(whcompat.Context(r), key, val(r))))
	})
}

func init() {
	rng := rand.New(rand.NewSource(time.Now().Unix()))
	idCounter = uint64(rng.Int63())
	inc = uint64(rng.Int63() | 3)
}

func newId() int64 {
	id := atomic.AddUint64(&idCounter, inc)
	return int64(id >> 1)
}
