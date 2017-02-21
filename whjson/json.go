// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

// Package whjson provides some nice utilities for dealing with JSON-based
// APIs, such as a good JSON wherr.Handler.
package whjson // import "gopkg.in/webhelp.v1/whjson"

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/spacemonkeygo/errors/errhttp"
	"gopkg.in/webhelp.v1/whcompat"
	"gopkg.in/webhelp.v1/wherr"
)

var (
	// ErrHandler provides a good wherr.Handler. It will return a JSON object
	// like `{"err": "message"}` where message is filled in with
	// errhttp.GetErrorBody. The status code is set with errhttp.GetStatusCode.
	ErrHandler = wherr.HandlerFunc(errHandler)
)

func errHandler(w http.ResponseWriter, r *http.Request, handledErr error) {
	log.Printf("error: %v", handledErr)
	data, err := json.MarshalIndent(map[string]string{
		"err": errhttp.GetErrorBody(handledErr)}, "", "  ")
	if err != nil {
		log.Printf("failed serializing error: %v", handledErr)
		data = []byte(`{"err": "Internal Server Error"}`)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", fmt.Sprint(len(data)))
	w.WriteHeader(errhttp.GetStatusCode(handledErr,
		http.StatusInternalServerError))
	w.Write(data)
}

// Render will render JSON `value` like `{"resp": <value>}`, falling back to
// ErrHandler if no error handler was registered and an error is
// encountered. This is good for making sure your API is always returning
// usefully namespaced JSON objects that are clearly differentiated from error
// responses.
func Render(w http.ResponseWriter, r *http.Request, value interface{}) {
	data, err := json.MarshalIndent(
		map[string]interface{}{"resp": value}, "", "  ")
	if err != nil {
		if handler := wherr.HandlingWith(whcompat.Context(r)); handler != nil {
			handler.HandleError(w, r, err)
			return
		}
		errHandler(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", fmt.Sprint(len(data)))
	w.Write(data)
}
