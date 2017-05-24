// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

// Package whparse provides some convenient input parsing helpers.
package whparse // import "gopkg.in/webhelp.v1/whparse"

import (
	"strconv"
	"strings"
)

// ParseBool is like strconv.ParseBool but also understands yes/no
func ParseBool(val string) (bool, error) {
	val = strings.ToLower(val)
	switch val {
	case "no", "n":
		return false, nil
	case "yes", "y":
		return true, nil
	}
	return strconv.ParseBool(val)
}

// OptInt64 parses val and returns the integer value in question, unless
// parsing fails, in which case the default def is returned.
func OptInt64(val string, def int64) int64 {
	if val == "" {
		return def
	}
	rv, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return def
	}
	return rv
}

// OptUint64 parses val and returns the integer value in question, unless
// parsing fails, in which case the default def is returned.
func OptUint64(val string, def uint64) uint64 {
	if val == "" {
		return def
	}
	rv, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return def
	}
	return rv
}

// OptInt32 parses val and returns the integer value in question, unless
// parsing fails, in which case the default def is returned.
func OptInt32(val string, def int32) int32 {
	if val == "" {
		return def
	}
	rv, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		return def
	}
	return int32(rv)
}

// OptUint32 parses val and returns the integer value in question, unless
// parsing fails, in which case the default def is returned.
func OptUint32(val string, def uint32) uint32 {
	if val == "" {
		return def
	}
	rv, err := strconv.ParseUint(val, 10, 32)
	if err != nil {
		return def
	}
	return uint32(rv)
}

// OptInt parses val and returns the integer value in question, unless
// parsing fails, in which case the default def is returned.
func OptInt(val string, def int) int {
	if val == "" {
		return def
	}
	rv, err := strconv.ParseInt(val, 10, 0)
	if err != nil {
		return def
	}
	return int(rv)
}

// OptUint parses val and returns the integer value in question, unless
// parsing fails, in which case the default def is returned.
func OptUint(val string, def uint) uint {
	if val == "" {
		return def
	}
	rv, err := strconv.ParseUint(val, 10, 0)
	if err != nil {
		return def
	}
	return uint(rv)
}

// OptBool parses val and returns the bool value in question, unless
// parsing fails, in which case the default def is returned.
func OptBool(val string, def bool) bool {
	if val == "" {
		return def
	}
	rv, err := ParseBool(val)
	if err != nil {
		return def
	}
	return rv
}

// OptFloat64 parses val and returns the float value in question, unless
// parsing fails, in which case the default def is returned.
func OptFloat64(val string, def float64) float64 {
	if val == "" {
		return def
	}
	rv, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return def
	}
	return rv
}

// OptFloat32 parses val and returns the float value in question, unless
// parsing fails, in which case the default def is returned.
func OptFloat32(val string, def float32) float32 {
	if val == "" {
		return def
	}
	rv, err := strconv.ParseFloat(val, 32)
	if err != nil {
		return def
	}
	return float32(rv)
}
