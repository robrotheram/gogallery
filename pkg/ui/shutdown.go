package ui

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const gracefulShutdownTimeout = 3 * time.Second

func installShutdownHandler(quit func(), stopped <-chan struct{}) func() {
	signals := make(chan os.Signal, 2)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	go listenForShutdown(signals, stopped, quit, os.Exit, gracefulShutdownTimeout)

	return func() {
		signal.Stop(signals)
	}
}

func listenForShutdown(
	signals <-chan os.Signal,
	stopped <-chan struct{},
	quit func(),
	forceExit func(int),
	timeout time.Duration,
) {
	var received os.Signal
	select {
	case received = <-signals:
		log.Printf("Received %s; shutting down", received)
	case <-stopped:
		return
	}

	// Fyne normally queues its signal-triggered quit on the UI event loop. Calling
	// the driver's thread-safe Quit directly also works when that queue is busy.
	go quit()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-stopped:
		return
	case repeated := <-signals:
		log.Printf("Received %s during shutdown; forcing exit", repeated)
		forceExit(signalExitCode(repeated))
	case <-timer.C:
		log.Printf("Graceful shutdown timed out; forcing exit")
		forceExit(signalExitCode(received))
	}
}

func signalExitCode(sig os.Signal) int {
	if number, ok := sig.(syscall.Signal); ok {
		return 128 + int(number)
	}
	return 1
}
