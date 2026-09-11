package task

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"

	cd "github.com/muidea/magicCommon/def"
)

func TestSyncSubmissionFailuresAreReturned(t *testing.T) {
	q := NewBackgroundRoutine(1)
	if !q.Shutdown(context.Background()) {
		t.Fatal("shutdown failed")
	}
	for _, err := range []error{
		q.SyncTask(&routineTask{funcPtr: func() { t.Error("rejected task ran") }}),
		q.SyncFunction(func() { t.Error("rejected function ran") }),
	} {
		if err == nil {
			t.Fatal("closed queue reported success")
		}
	}
}

func TestSyncInvalidArgumentsDoNotExecute(t *testing.T) {
	q := NewBackgroundRoutine(1)
	defer q.Shutdown(context.Background())
	for _, err := range []error{
		q.SyncTask(nil), q.SyncTaskWithTimeOut(nil, time.Second),
		q.SyncFunction(nil), q.SyncFunctionWithTimeOut(nil, time.Second),
		q.SyncFunctionWithTimeOut(func() { t.Error("invalid timeout task ran") }, -2),
	} {
		if err == nil {
			t.Fatal("invalid argument reported success")
		}
	}
}

func TestSyncPanicCompletesWaitAndWorkerRemainsUsable(t *testing.T) {
	q := NewBackgroundRoutine(1)
	defer q.Shutdown(context.Background())
	err := q.SyncFunction(func() { panic("task failed") })
	var failure *cd.Error
	if !errors.As(err, &failure) || failure.Code != cd.Unexpected {
		t.Fatalf("panic result: %v", err)
	}
	if err := q.SyncFunction(func() {}); err != nil {
		t.Fatal(err)
	}
}

func TestSyncTimeoutDoesNotCancelLateCompletion(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		q := NewBackgroundRoutine(1)
		release := make(chan struct{})
		finished := make(chan struct{})
		err := q.SyncFunctionWithTimeOut(func() { <-release; close(finished) }, time.Second)
		var failure *cd.Error
		if !errors.As(err, &failure) || failure.Code != cd.Timeout {
			t.Fatalf("timeout result: %v", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if q.Shutdown(ctx) {
			t.Fatal("timeout was mistaken for actual completion")
		}
		close(release)
		if !q.Shutdown(context.Background()) {
			t.Fatal("late completion failed to drain")
		}
		<-finished
	})
}

func TestSyncSuccessAndAlreadyPublishedCompletion(t *testing.T) {
	q := NewBackgroundRoutine(1)
	defer q.Shutdown(context.Background())
	if err := q.SyncFunctionWithTimeOut(func() {}, time.Second); err != nil {
		t.Fatal(err)
	}
	s := &syncTask{rawTask: &routineTask{funcPtr: func() {}}, resultChannel: make(chan error, 1)}
	s.Run()
	if err := s.Wait(0); err != nil {
		t.Fatalf("completed task reported timeout: %v", err)
	}
}
