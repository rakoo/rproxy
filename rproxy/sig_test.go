package rproxy

import (
	"bytes"
	"os"
	"testing"
)

func TestSignature(t *testing.T) {
	in, err := os.Open("Obama")
	if err != nil {
		t.Fatal("Couldn't open \"Obama\" file for calculating sigs:", err)
	}

	stat, err := in.Stat()
	if err != nil {
		t.Fatal("Couldn't stat file:", err)
	}
	sig := Signature(in, uint64(stat.Size()))
	if err != nil {
		t.Fatal(err)
	}

	magic := sig[0:6]
	if string(magic) != "rproxy" {
		t.Fatal("Magic string is not \"rproxy\" but", string(magic))
	}

	version := sig[6:10]
	if bytes.Compare(version, []byte{0, 0, 0, 1}) != 0 {
		t.Fatal("Version is incorrect: got", version)
	}

	blocksize := sig[10:14]
	if bytes.Compare(blocksize, []byte{0, 0, 91, 199}) != 0 {
		t.Fatal("Blocksize is incorrect: expected 23495, got", blocksize)
	}
}