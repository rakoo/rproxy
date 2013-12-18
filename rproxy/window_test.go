package rproxy

import (
	"testing"
)

func TestFewAppends(t *testing.T) {
	w := NewWindow(1024)
	for _, x := range []byte("this is a test message") {
		w.Append(x)
	}

	p := w.Bytes()

	if w.Filled != 22 {
		t.Errorf("We should have filled 22 bytes, got", w.Filled)
	}
	if string(p) != "this is a test message" {
		t.Errorf("Wrong bytes: got \"%s\"", string(p))
	}
}

func TestLotAppends(t *testing.T) {
	w := NewWindow(22)
	for i := 0; i < 1000; i++ {
		w.Append(byte(i))
	}

	for _, x := range []byte("this is a test message") {
		w.Append(x)
	}

	p := w.Bytes()

	if w.Filled != 1022 {
		t.Errorf("We should have filled 1022 bytes, got", w.Filled)
	}
	if string(p) != "this is a test message" {
		t.Errorf("Wrong bytes: got \"%s\"", string(p))
	}
}

func TestReset(t *testing.T) {
	w := NewWindow(10)
	for i := 0; i < 1000; i++ {
		w.Append(byte(i))
	}

	w.Reset()

	if len(w.Bytes()) != 0 {
		t.Error("There should be nothing after Reset() !")
	}
	if w.Filled != 0 {
		t.Error("There should be nothing filled after reset")
	}
}
