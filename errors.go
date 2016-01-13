package webhelp

import (
	"net/http"

	"github.com/spacemonkeygo/errors"
	"github.com/spacemonkeygo/errors/errhttp"
)

var (
	HTTPError   = errors.NewClass("HTTP Error", errors.NoCaptureStack())
	ErrNotFound = HTTPError.NewClass("Not found",
		errhttp.SetStatusCode(http.StatusNotFound))
	ErrMethodNotAllowed = HTTPError.NewClass("Method not allowed",
		errhttp.SetStatusCode(http.StatusMethodNotAllowed))
)
