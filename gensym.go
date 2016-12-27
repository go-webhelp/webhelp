// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package webhelp

import (
	"sync"
)

var (
	keyMtx     sync.Mutex
	keyCounter uint64
)

// ContextKey is only useful via the GenSym() constructor. See GenSym() for
// more documentation
type ContextKey struct {
	id uint64
}

// GenSym generates a brand new, never-before-seen ContextKey for use as a
// Context.WithValue key. Please see the example.
func GenSym() ContextKey {
	keyMtx.Lock()
	defer keyMtx.Unlock()
	keyCounter += 1
	return ContextKey{id: keyCounter}
}
