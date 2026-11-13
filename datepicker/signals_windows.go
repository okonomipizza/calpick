//go:build windows

package datepicker

import (
	"os"
	"os/signal"
	"syscall"
)

// watchSignals makes Ctrl-Break and console close events cancel the picker so
// its deferred terminal restore can run. Ctrl-C is normally delivered as a
// console key event because processed input is disabled.
func watchSignals() func() {
	ch := make(chan os.Signal, 8)
	done := make(chan struct{})
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-ch:
				interruptRequested.Store(true)
			case <-done:
				return
			}
		}
	}()

	return func() {
		signal.Stop(ch)
		close(done)
	}
}
