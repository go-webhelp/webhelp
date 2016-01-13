// package webhelp is a bunch of useful utilities for whenever I do web
// programming in Go. Like a framework, but better, cause it's not.
//
// Note that this tightly integrates with Context objects. You can read more
// about them here: https://blog.golang.org/context
//
// Especially important if you rely on context cancelation is that in this
// library, incoming requests receive Context objects that are canceled if
// the incoming request is on a transport that supports close notification.
package webhelp
