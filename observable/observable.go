package observable

import (
	"sync"
	"time"
)

const DefaultTimeout = 5 * time.Second

type ID uint64

type SendOption struct {
	Timeout time.Duration
}

type Subscription[T any] struct {
	Channel    chan T
	observable *Observable[T]
}

func (s *Subscription[T]) Observe() <-chan T {
	return s.Channel
}

func (s *Subscription[T]) Unsubscribe() {
	s.observable.Unsubscribe(s)
}

type Observable[T any] struct {
	subscriptions []*Subscription[T]
	locker        sync.Locker
}

func (o *Observable[T]) Subscribe() *Subscription[T] {
	o.locker.Lock()
	defer o.locker.Unlock()

	subscription := &Subscription[T]{
		Channel:    make(chan T),
		observable: o,
	}

	o.subscriptions = append(o.subscriptions, subscription)

	return subscription
}

func (o *Observable[T]) Unsubscribe(subscription *Subscription[T]) {
	o.locker.Lock()
	defer o.locker.Unlock()

	for i, s := range o.subscriptions {
		if s == subscription {
			o.subscriptions = append(o.subscriptions[:i], o.subscriptions[i+1:]...)
			close(s.Channel)
			break
		}
	}
}

func (o *Observable[T]) Send(data T, options ...SendOption) bool {
	o.locker.Lock()
	defer o.locker.Unlock()

	if len(o.subscriptions) == 0 {
		return false
	}

	timeout := DefaultTimeout

	for _, option := range options {
		if option.Timeout > 0 {
			timeout = option.Timeout
		}
	}

	for _, subscription := range o.subscriptions {
		go func(subscription *Subscription[T]) {
			select {
			case subscription.Channel <- data:
			case <-time.After(timeout):
				subscription.Unsubscribe()
			}
		}(subscription)
	}

	return true
}

func (o *Observable[T]) SubscriptionCount() int {
	o.locker.Lock()
	defer o.locker.Unlock()
	return len(o.subscriptions)
}

func New[T any]() *Observable[T] {
	return &Observable[T]{
		locker: &sync.Mutex{},
	}
}

type Group[T any] struct {
	observables map[ID]*Observable[T]
	locker      sync.Locker
}

func (r *Group[T]) Send(id ID, data T, options ...SendOption) bool {
	r.locker.Lock()
	defer r.locker.Unlock()

	observable, ok := r.observables[id]
	if !ok {
		return false
	}

	return observable.Send(data, options...)
}

func (r *Group[T]) Subscribe(id ID) *Subscription[T] {
	r.locker.Lock()
	defer r.locker.Unlock()

	observable, ok := r.observables[id]
	if !ok {
		observable = New[T]()
		r.observables[id] = observable
	}

	return observable.Subscribe()
}

func (r *Group[T]) SubscriptionCount(id ID) (int, bool) {
	r.locker.Lock()
	defer r.locker.Unlock()

	if observable, ok := r.observables[id]; ok {
		return observable.SubscriptionCount(), true
	}

	return 0, false
}

func NewGroup[T any]() *Group[T] {
	return &Group[T]{
		observables: make(map[ID]*Observable[T]),
		locker:      &sync.Mutex{},
	}
}
