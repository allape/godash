package observable

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestObservable(t *testing.T) {
	GroupId := ID(rand.Intn(1000))
	MagicWord := rand.Intn(10000)

	obs := NewGroup[int]()

	var waitGroup sync.WaitGroup

	sub1 := obs.Subscribe(GroupId)
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		for i := range sub1.Observe() {
			if i != MagicWord {
				t.Errorf("expected %d, got %d", MagicWord, i)
			}
			sub1.Unsubscribe()
		}
	}()

	sub2 := obs.Subscribe(GroupId)
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		for i := range sub2.Observe() {
			if i != MagicWord {
				t.Errorf("expected %d, got %d", MagicWord, i)
			}
			sub2.Unsubscribe()
		}
	}()

	obs.Send(GroupId, MagicWord, SendOption{
		Timeout: time.Second,
	})

	done := make(chan struct{}, 1)

	go func() {
		waitGroup.Wait()
		done <- struct{}{}
	}()

	select {
	case <-done:
		close(done)
	case <-time.After(30 * time.Second):
		t.Fatalf("timeout")
	}

	if count, ok := obs.SubscriptionCount(GroupId); ok {
		if count != 0 {
			t.Errorf("expected 0, got %d", count)
		}
	} else {
		t.Error("expected true, got false")
	}
}

func TestTimeout(t *testing.T) {
	GroupId := ID(rand.Intn(1000))
	MagicWord := rand.Intn(10000)

	obs := NewGroup[int]()

	sub1 := obs.Subscribe(GroupId)
	go func() {
		for i := range sub1.Observe() {
			if i != MagicWord {
				t.Errorf("expected %d, got %d", MagicWord, i)
			}
		}
	}()
	sub2 := obs.Subscribe(GroupId)
	go func() {
		for range sub2.Observe() {
			for i := range sub2.Observe() {
				if i != MagicWord {
					t.Errorf("expected %d, got %d", MagicWord, i)
				}
			}
		}
	}()
	obs.Subscribe(GroupId)

	if count, ok := obs.SubscriptionCount(GroupId); ok {
		if count != 3 {
			t.Errorf("expected 3, got %d", count)
		}
	} else {
		t.Error("expected true, got false")
	}

	obs.Send(GroupId, MagicWord, SendOption{
		Timeout: time.Second,
	})

	time.Sleep(3 * time.Second)

	if count, ok := obs.SubscriptionCount(GroupId); ok {
		if count != 2 {
			t.Errorf("expected 2, got %d", count)
		}
	} else {
		t.Error("expected true, got false")
	}

	sub1.Unsubscribe()
	sub2.Unsubscribe()

	if count, ok := obs.SubscriptionCount(GroupId); ok {
		if count != 0 {
			t.Errorf("expected 0, got %d", count)
		}
	} else {
		t.Error("expected true, got false")
	}
}
