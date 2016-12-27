// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

// Package whcompat provides webhelp compatibility across different Go
// releases.
//
// The webhelp suite depends heavily on Go 1.7 style http.Request contexts,
// which aren't available in earlier Go releases. This package backports all
// of the functionality in a forwards-compatible way. You can use this package
// to get the desired behavior for all Go releases.
package whcompat
