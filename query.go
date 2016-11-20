// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package webhelp

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
