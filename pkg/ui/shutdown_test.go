package ui

import (
	"os"
	"syscall"
	"testing"
	"time"
)

func TestListenForShutdownQuitsGracefully(t *testing.T) {
	signals := make(chan os.Signal, 1)
	stopped := make(chan struct{})
	quitCalled := make(chan struct{})
	forced := make(chan int, 1)

	go listenForShutdown(
		signals,
		stopped,
		func() { close(quitCalled) },
		func(code int) { forced <- code },
		time.Second,
	)

	signals <- os.Interrupt
	select {
	case <-quitCalled:
	case <-time.After(time.Second):
		t.Fatal("quit was not called")
	}
	close(stopped)

	select {
	case code := <-forced:
		t.Fatalf("force exit called with code %d", code)
	case <-time.After(20 * time.Millisecond):
	}
}

func TestListenForShutdownForcesARepeatedInterrupt(t *testing.T) {
	signals := make(chan os.Signal, 2)
	stopped := make(chan struct{})
	quitCalled := make(chan struct{})
	forced := make(chan int, 1)

	go listenForShutdown(
		signals,
		stopped,
		func() { close(quitCalled) },
		func(code int) { forced <- code },
		time.Second,
	)

	signals <- os.Interrupt
	select {
	case <-quitCalled:
	case <-time.After(time.Second):
		t.Fatal("quit was not called")
	}
	signals <- os.Interrupt

	select {
	case code := <-forced:
		if code != 128+int(syscall.SIGINT) {
			t.Fatalf("unexpected exit code: got %d", code)
		}
	case <-time.After(time.Second):
		t.Fatal("force exit was not called")
	}
}

func TestListenForShutdownForcesExitAfterTimeout(t *testing.T) {
	signals := make(chan os.Signal, 1)
	stopped := make(chan struct{})
	forced := make(chan int, 1)

	go listenForShutdown(
		signals,
		stopped,
		func() {},
		func(code int) { forced <- code },
		10*time.Millisecond,
	)

	signals <- syscall.SIGTERM
	select {
	case code := <-forced:
		if code != 128+int(syscall.SIGTERM) {
			t.Fatalf("unexpected exit code: got %d", code)
		}
	case <-time.After(time.Second):
		t.Fatal("force exit was not called")
	}
}
