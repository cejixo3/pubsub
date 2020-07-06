package pubsub

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	lib := New()
	if lib == nil {
		t.FailNow()
	}
}

func TestPubSub_Unsubscribe(t *testing.T) {
	defer func() {
		if p := recover(); p != nil {
			t.FailNow()
		}
	}()
	lib := New()
	tn := "some topic"
	sn := "subscriber/id"
	sn2 := "subscriber not exist id"
	lib.Subscribe(tn, sn)
	lib.Publish(tn, []byte("message"))
	lib.Unsubscribe(tn, sn)
	lib.Unsubscribe(tn, sn2)
	if _, err := lib.Poll(tn, sn); err != ErrNoSubscriptions {
		t.FailNow()
	}
	if _, err := lib.Poll(tn, sn2); err != ErrNoSubscriptions {
		t.FailNow()
	}
}

func TestPubSub_Subscribe_Publish_Poll(t *testing.T) {
	defer func() {
		if p := recover(); p != nil {
			t.FailNow()
		}
	}()
	lib := New()
	tn := "some topic"
	sn := "subscriber/id"
	lib.Subscribe(tn, sn)
	lib.Publish(tn, []byte("message"))
	msg, _ := lib.Poll(tn, sn)
	if msg == nil {
		t.FailNow()
	}
}

func TestPubSub_PollParallel(t *testing.T) {
	defer func() {
		if p := recover(); p != nil {
			t.FailNow()
		}
	}()
	tn := "testing topic name"
	tn2 := tn + " 2"
	snf := "testing subscriber name format %d"
	topics := []string{tn, tn2}
	lib := New()
	nPol := 4
	var wg sync.WaitGroup
	sentSeq := "sequence for sending"
	allSend := false
	wg.Add(1 + nPol*len(topics))
	for i := 0; i < nPol; i++ {
		lib.Subscribe(tn, fmt.Sprintf(snf, i))
		lib.Subscribe(tn2, fmt.Sprintf(snf, i))
	}
	go func() {
		for _, tn := range topics {
			for i := 0; i < len([]byte(sentSeq)); i++ {
				lib.Publish(tn, []byte{[]byte(sentSeq)[i]})
			}
		}
		allSend = true
		wg.Done()
	}()
	time.Sleep(1 * time.Second)
	for _, tn := range topics {
		for i := 0; i < nPol; i++ {
			go func(n int, tn string) {
				var seq []byte
				for {
					time.Sleep(1 * time.Millisecond)
					msg, err := lib.Poll(tn, fmt.Sprintf(snf, n))
					if err == ErrNoSubscriptions {
						t.Errorf("error in subscribe mechanism #%d", n)
						t.FailNow()
					}
					if len(msg) == 0 && allSend {
						break
					}
					seq = append(seq, msg[0])
				}
				wg.Done()
				if sentSeq != string(seq) {
					t.Errorf("broken order in poller #%d", n)
					t.FailNow()
				}
			}(i, tn)
		}
	}
	wg.Wait()
}
