package pushover

import (
	"log"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	load(t)
}

func TestHasAppAndKey(t *testing.T) {
	p := load(t)
	if !p.HasApp("a1", "a2") {
		t.Errorf("cannot find application keys a1 and a2")
	}
	if p.HasApp("a1", "a2", "a3") {
		t.Errorf("found bad application key 'a3'")
	}
	if p.HasApp("not there") {
		t.Errorf("found application key 'not there'")
	}
	if p.HasApp("r1") {
		t.Errorf("found application key 'r1'")
	}
	if p.HasRec("a1") {
		t.Errorf("found receiver key 'a1'")
	}
	if !p.HasRec("r1", "r2") {
		t.Errorf("cannot find receiver keys r1 and r2")
	}
}

func TestMessage(t *testing.T) {
	_ = message(t)
}

func TestThrottle(t *testing.T) {
	m := message(t)
	var counter int

	m.Throttle(0)
	counter = 0
	count := func() error { counter++; return nil }
	for i := 0; i < 10; i++ {
		if err := m.runThrottled(count); err != nil {
			t.Errorf("runner returned error, throttle=%s, err=%s", m.throttle, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	if counter != 10 {
		t.Errorf("throttled, count=%d, want 10", counter)
	}

	m.Throttle(time.Second)
	counter = 0
	m.runThrottled(count)
	for i := 0; i < 10; i++ {
		switch err := m.runThrottled(count); {
		case err == nil:
			t.Error("got NIL error, should have throttled")
		case err != ErrThrottled:
			t.Errorf("runner returned error, throttle=%s, err=%s", m.throttle, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	if counter != 1 {
		t.Errorf("not throttled, count=%d, want 10", counter)
	}

}

func message(t *testing.T) Message {
	p := load(t)
	m, err := p.Message("a1", "r1")
	if err != nil {
		t.Fatalf("cannot create message a1/r1")
	}
	return m
}

func load(t *testing.T) Pushover {
	const fname = "sample.json"
	p, err := Load(fname)
	if err != nil {
		log.Fatalf("cannot load config file %s", fname)
	}
	return p
}
