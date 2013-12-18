package rproxy

import (
	"hash/adler32"
	"math/rand"
	"testing"
)

func TestChecksumEqualsRollLittleInput(t *testing.T) {
	in := make([]byte, 1000)
	for i := 0; i < len(in); i++ {
		in[i] = byte(rand.Intn(255))
	}

	cs := Checksum(in)
	rs := NewRollsum(len(in))
	for _, x := range in {
		rs.Roll(0, x)
	}

	if cs != rs.Digest() {
		t.Fatalf("Rollsum differs on input smaller than window: expected %d, got %d", cs, rs.Digest())
	}
}

func TestChecksumEqualsRoll(t *testing.T) {
	in := make([]byte, 10000)
	for i := 0; i < len(in); i++ {
		in[i] = byte(rand.Intn(255))
	}

	window := 1024

	cs := Checksum(in[0:window])
	rs := NewRollsum(window)
	for _, x := range in[0:window] {
		rs.Roll(0, x)
	}
	ds := rs.Digest()

	if cs != ds {
		t.Errorf("Error initial: got %d and %d", cs, ds)
	}

	for idx := 1; idx < len(in)-window; idx++ {
		cs = Checksum(in[idx : idx+window])
		rs.Roll(in[idx-1], in[idx+window-1])

		if cs != rs.Digest() {
			t.Errorf("Error at %d: expected %d, got %d", idx, cs, rs.Digest())
		}
	}
}

func BenchmarkRolling10Elements(b *testing.B) {
	in := make([]byte, 10)
	for i := 0; i < len(in); i++ {
		in[i] = byte(rand.Intn(255))
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rs := NewRollsum(len(in))
		for _, x := range in {
			rs.Roll(0, x)
		}
		rs.Digest()
	}
	b.SetBytes(int64(len(in)))
}

func BenchmarkRolling100Elements(b *testing.B) {
	in := make([]byte, 100)
	for i := 0; i < len(in); i++ {
		in[i] = byte(rand.Intn(255))
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rs := NewRollsum(len(in))
		for _, x := range in {
			rs.Roll(0, x)
		}
		rs.Digest()
	}
	b.SetBytes(int64(len(in)))
}

func BenchmarkRolling1000Elements(b *testing.B) {
	in := make([]byte, 1000)
	for i := 0; i < len(in); i++ {
		in[i] = byte(rand.Intn(255))
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rs := NewRollsum(len(in))
		for _, x := range in {
			rs.Roll(0, x)
		}
		rs.Digest()
	}
	b.SetBytes(int64(len(in)))
}
func BenchmarkRolling10000Elements(b *testing.B) {
	in := make([]byte, 10000)
	for i := 0; i < len(in); i++ {
		in[i] = byte(rand.Intn(255))
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rs := NewRollsum(len(in))
		for _, x := range in {
			rs.Roll(0, x)
		}
		rs.Digest()
	}
	b.SetBytes(int64(len(in)))
}

func BenchmarkChecksum10000Elements(b *testing.B) {
	in := make([]byte, 10000)
	for i := 0; i < len(in); i++ {
		in[i] = byte(rand.Intn(255))
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Checksum(in)
	}
	b.SetBytes(int64(len(in)))
}

func BenchmarkAdler32(b *testing.B) {
	in := make([]byte, 10000)
	for i := 0; i < len(in); i++ {
		in[i] = byte(rand.Intn(255))
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		adler32.Checksum(in)
	}
	b.SetBytes(int64(len(in)))
}
