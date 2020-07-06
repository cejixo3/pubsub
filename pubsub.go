/*
	Simple in-memory implementation of Pub/Sub with polling an approach. You can use this package for building pub/sub
	systems where the main method of obtaining data is poling (like cases with http). Messages are saved until the
	subscriber picks them up. This package uses []byte as "message format".

	Each subscription stores an slice of pointers to messages (no copy - just pointers).
	Storage complexity: messages: O(n) + pointers: O(k*n) where
	n - number of messages,
	k - number of subscribers
	Also topic name and subscriber name must take part in complexity calculation
*/
package pubsub

import (
	"errors"
	"sync"
)

// Error happens only if subscription not exist already
var ErrNoSubscriptions = errors.New("there are no subscriptions")

// Storage for messages (something like FIFO stack)
type sliceStorage [][]byte

// Add message (b) to the end of slice
func (s *sliceStorage) add(b []byte) {
	*s = append(*s, b)
}

// Take a "oldest" message from slice and remove it from slice
func (s *sliceStorage) take() []byte {
	if len(*s) > 0 {
		msg := (*s)[0]
		*s = (*s)[1:]
		return msg
	}
	return nil
}

// List of subscriptions protected by mutex
// hm - hashmap where key is subscription name (sn) and value - a list of messages for this subscription name
type subscriptions struct {
	mux sync.Mutex
	hm  map[string]*sliceStorage
}

// PubSuber interface helps to hide `pubSub` from direct access/initialization and make ability to
// pass instance of PubSuber into another function, declare variables like: var br pubsub.PubSuber, etc
type PubSuber interface {
	// Publish message
	Publish(tn string, b []byte)
	// Subscribe for messages by topic and subscription name
	Subscribe(tn, sn string)
	// Unsubscribe for messages by topic and subscription name
	Unsubscribe(tn, sn string)
	// Fetching messages for topic name (tn) and subscriber name (sn)
	Poll(tn, sn string) ([]byte, error)
}

// List of subscriptions protected by RW mutex
// RW mutex used because access to `hm` not always means write operations
type pubSub struct {
	mux sync.RWMutex
	hm  map[string]*subscriptions
}

// Publish message (b) by topic name (tn) if have subscriptions already
// Complexity: O(N+1)
func (p *pubSub) Publish(tn string, b []byte) {
	p.mux.RLock()
	subs, ok := p.hm[tn]
	p.mux.RUnlock()
	if ok {
		subs.mux.Lock()
		for _, sub := range subs.hm {
			sub.add(b)
		}
		subs.mux.Unlock()
	}
}

// Subscribe to message by topic name (tn) and subscriber name (sn)
// Creates new topic if not exist before
func (p *pubSub) Subscribe(tn, sn string) {
	p.mux.Lock()
	defer p.mux.Unlock()
	if subs, ok := p.hm[tn]; !ok {
		p.hm[tn] = &subscriptions{
			hm: map[string]*sliceStorage{sn: &sliceStorage{}},
		}
	} else {
		subs.mux.Lock()
		if _, ok := subs.hm[sn]; !ok {
			subs.hm[sn] = &sliceStorage{}
		}
		subs.mux.Unlock()
	}
}

// Unsubscribe by topic name (tn) and subscriber name (sn)
// @todo implement removing keys from p.hm[tn] when subscription list is empty
func (p *pubSub) Unsubscribe(tn, sn string) {
	p.mux.RLock()
	subs, ok := p.hm[tn]
	p.mux.RUnlock()
	if ok {
		subs.mux.Lock()
		delete(subs.hm, sn)
		subs.mux.Unlock()
	}
}

// Fetching messages for topic name (tn) and subscriber name (sn)
// error raises if no subscriptions
// nil, nil should be returned if all messages was fetched already
// Complexity: O(3)
func (p *pubSub) Poll(tn, sn string) ([]byte, error) {
	p.mux.RLock()
	subs, ok := p.hm[tn]
	p.mux.RUnlock()
	if !ok {
		return nil, ErrNoSubscriptions
	} else {
		subs.mux.Lock()
		defer subs.mux.Unlock()
		if sub, ok := subs.hm[sn]; ok {
			return sub.take(), nil
		} else {
			return nil, ErrNoSubscriptions
		}
	}
}

// Constructor. Creates an instance of PubSuber
// Using a PubSuber interface instead of a pointer to pubSub guarantees the using of this constructor in other packages
func New() PubSuber {
	return &pubSub{
		hm: map[string]*subscriptions{},
	}
}
