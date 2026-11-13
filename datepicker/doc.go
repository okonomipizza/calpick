// Package datepicker renders an interactive, clickable calendar on the
// controlling terminal and returns the selected date.
//
// The package uses only the Go standard library. Pick requires an interactive
// terminal; it draws on the controlling terminal (not stdout) and restores the
// terminal state before returning. A user can select a date with the mouse or
// the keyboard, change months with "<"/">" or PageUp/PageDown, and cancel with
// q, Esc, or Ctrl-C (ErrCanceled).
//
// Pick is not safe for concurrent use.
package datepicker
