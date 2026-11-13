//go:build linux || darwin

package datepicker

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func watchSignals() func() {
	ch := make(chan os.Signal, 8)
	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)

	signal.Notify(ch, syscall.SIGWINCH, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		defer wg.Done()
		for {
			select {
			case sig := <-ch:
				if sig == syscall.SIGWINCH {
					resizeRequested.Store(true)
				} else {
					interruptRequested.Store(true)
				}
			case <-done:
				return
			}
		}
	}()

	return func() {
		signal.Stop(ch)
		close(done)
		// Do not leave a goroutine that can set flags after Pick returns.
		wg.Wait()
	}
}
