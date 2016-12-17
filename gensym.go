// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package webhelp

import (
	"github.com/spacemonkeygo/errors"
)

// GenSym generates a brand new, never-before-seen symbol for use as a
// Context.WithValue key.
//
// Example usage:
//
//   var UserKey = webhelp.GenSym()
//
//   func myWrapper(h http.Handler) http.Handler {
//     return webhelp.RouteHandlerFunc(h,
//       func(w http.ResponseWriter, r *http.Request) {
//         user, err := loadUser(r)
//         if err != nil {
//           webhelp.HandleError(w, r, err)
//           return
//         }
//         h.ServeHTTP(w, webhelp.WithContext(r,
//             context.WithValue(webhelp.Context(r), UserKey, user)))
//       })
//   }
//
//   func myHandler(w http.ResponseWriter, r *http.Request) {
//     ctx := webhelp.Context(r)
//     if user, ok := ctx.Value(UserKey).(*User); ok {
//       // do something with the user
//     }
//   }
//
//   func Routes() http.Handler {
//     return myWrapper(http.HandlerFunc(myHandler))
//   }
//
func GenSym() interface{} { return errors.GenSym() }
