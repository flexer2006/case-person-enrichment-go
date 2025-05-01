package shutdown_test

import (
	"context"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/flexer2006/case-person-enrichment-go/pkg/shutdown"
)

const (
	errFailedToFindProcess        = "failed to find process:"
	errFailedToSendSignal         = "failed to send signal:"
	errHook1NotCalled             = "hook 1 was not called"
	errHook2NotCalled             = "hook 2 was not called"
	errWaitFunctionTimeout        = "wait function didn't return within the expected time"
	errWaitDidntRespectTimeout    = "wait didn't respect timeout: took"
	errSlowHookShouldntComplete   = "the slow hook shouldn't have completed"
	errHooksRunSequentially       = "hooks appear to run sequentially:"
	errTimeoutWaitingForHooks     = "timed out waiting for hooks to complete"
	errHookNotCalledAfterCancel   = "hook was not called after context cancellation"
	errWaitNotReturnedAfterCancel = "wait function didn't return after context cancellation"
)

func TestWaitExecutesHooks(t *testing.T) {
	hook1Called := make(chan struct{})
	hook2Called := make(chan struct{})

	hook1 := func(_ context.Context) error {
		close(hook1Called)
		return nil
	}

	hook2 := func(_ context.Context) error {
		close(hook2Called)
		return nil
	}

	ctx := context.Background()
	go func() {
		shutdown.Wait(ctx, time.Second, hook1, hook2)
	}()

	time.Sleep(100 * time.Millisecond)

	process, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("%s %v", errFailedToFindProcess, err)
	}
	if err := process.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("%s %v", errFailedToSendSignal, err)
	}

	select {
	case <-hook1Called:
	case <-time.After(2 * time.Second):
		t.Error(errHook1NotCalled)
	}

	select {
	case <-hook2Called:
	case <-time.After(2 * time.Second):
		t.Error(errHook2NotCalled)
	}
}

func TestWaitRespectsTimeout(t *testing.T) {
	var mtx sync.Mutex
	completed := false

	waitDone := make(chan struct{})

	slowHook := func(ctx context.Context) error {
		select {
		case <-time.After(2 * time.Second):
			mtx.Lock()
			completed = true
			mtx.Unlock()
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	start := time.Now()
	ctx := context.Background()
	go func() {
		shutdown.Wait(ctx, 500*time.Millisecond, slowHook)
		close(waitDone)
	}()

	time.Sleep(100 * time.Millisecond)
	process, _ := os.FindProcess(os.Getpid())
	_ = process.Signal(syscall.SIGTERM)

	select {
	case <-waitDone:
	case <-time.After(3 * time.Second):
		t.Fatal(errWaitFunctionTimeout)
	}

	elapsed := time.Since(start)
	if elapsed > 750*time.Millisecond {
		t.Errorf("%s %v", errWaitDidntRespectTimeout, elapsed)
	}

	mtx.Lock()
	defer mtx.Unlock()
	if completed {
		t.Error(errSlowHookShouldntComplete)
	}
}

func TestWaitRunsHooksConcurrently(t *testing.T) {
	var wgp sync.WaitGroup
	wgp.Add(2)

	start := time.Now()

	hook1 := func(_ context.Context) error {
		time.Sleep(500 * time.Millisecond)
		wgp.Done()
		return nil
	}

	hook2 := func(_ context.Context) error {
		time.Sleep(500 * time.Millisecond)
		wgp.Done()
		return nil
	}

	ctx := context.Background()
	go func() {
		shutdown.Wait(ctx, time.Second, hook1, hook2)
	}()

	time.Sleep(100 * time.Millisecond)
	process, _ := os.FindProcess(os.Getpid())
	_ = process.Signal(syscall.SIGTERM)

	waitCh := make(chan struct{})
	go func() {
		wgp.Wait()
		close(waitCh)
	}()

	select {
	case <-waitCh:
		elapsed := time.Since(start)
		if elapsed >= 900*time.Millisecond {
			t.Errorf("%s %v", errHooksRunSequentially, elapsed)
		}
	case <-time.After(2 * time.Second):
		t.Fatal(errTimeoutWaitingForHooks)
	}
}

func TestWaitContextCancellation(t *testing.T) {
	hook1Called := make(chan struct{})

	hook1 := func(_ context.Context) error {
		close(hook1Called)
		return nil
	}

	waitDone := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		shutdown.Wait(ctx, time.Second, hook1)
		close(waitDone)
	}()

	time.Sleep(100 * time.Millisecond)

	cancel()

	select {
	case <-hook1Called:
	case <-time.After(2 * time.Second):
		t.Error(errHookNotCalledAfterCancel)
	}

	select {
	case <-waitDone:
	case <-time.After(2 * time.Second):
		t.Fatal(errWaitNotReturnedAfterCancel)
	}
}
