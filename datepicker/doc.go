// Package datepicker renders an interactive, clickable calendar on the
// controlling terminal and returns the selected date or date range.
//
// The package uses only the Go standard library. Pick selects one date;
// PickRange selects From and Until on one calendar, showing the range below
// it. Both require an interactive terminal, draw on the controlling terminal
// (not stdout), and restore the terminal state before returning.
//
// A user can select dates with the mouse or the keyboard, change months with
// "<"/">" or PageUp/PageDown, select range endpoints with Space, and cancel
// with q, Esc, or Ctrl-C (ErrCanceled).
//
// Pick and PickRange are not safe for concurrent use.
package datepicker
