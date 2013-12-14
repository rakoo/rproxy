package rproxy

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"hash/adler32"
	"io"
	"log"
	"math"
)

// Simple, stupid differ: adler-32 as weak, sha1 as strong

const MAX_SIG_SIZE = 1024

// Calculate the signature of given reader
func Signature(rd io.Reader, inputSize uint64) (sig []byte) {
	blocksize := getBlockSize(inputSize)

	p := make([]byte, blocksize)
	lastblock := false

	hs := make([]*WeakStrongHash, 0)

	for {
		n, err := rd.Read(p)
		if err != nil {
			if err == io.EOF {
				if n == 0 {
					break
				} else {
					lastblock = true
				}
			} else {
				log.Println("Couldn't read from input:", err)
				return
			}
		}

		if lastblock {
			p = p[:n]
		}

		h := &WeakStrongHash{
			Weak:   adler32.Checksum(p),
			Strong: sha1sum(p),
		}
		hs = append(hs, h)

		if lastblock {
			break
		}

		p = make([]byte, blocksize)
	}

	return serializeSig(hs, blocksize)
}

func serializeSig(in []*WeakStrongHash, blocksize uint32) []byte {
	var out bytes.Buffer
	out.WriteString("rproxy")
	binary.Write(&out, binary.BigEndian, uint32(1)) // version
	binary.Write(&out, binary.BigEndian, blocksize)

	for _, hash := range in {
		binary.Write(&out, binary.BigEndian, hash.Weak)
		binary.Write(&out, binary.BigEndian, hash.Strong)
	}

	return out.Bytes()
}

// Get the size of the blocks depending on input size
// We don't want signature much larger than 1024B
func getBlockSize(inputsize uint64) uint32 {
	// Substract version and block size in available size
	maxSize := MAX_SIG_SIZE - 4 - 4

	maxBlocks := math.Ceil(float64(maxSize) / float64(adler32.Size+sha1.Size))
	return uint32(math.Ceil(float64(inputsize) / maxBlocks))
}

func readSig(sig []byte) (blocksize int, hashes []*WeakStrongHash) {

	rd := bytes.NewReader(sig)

	magic := make([]byte, 6)
	rd.Read(magic)
	if string(magic) != "rproxy" {
		log.Println("Wrong file format: magic is not \"rproxy\")")
		return
	}

	var version uint32
	if binary.Read(rd, binary.BigEndian, &version) != nil {
		return
	}

	if version != 1 {
		log.Println("Expected version 1, got", version)
		return
	}

	var blocksize32 uint32
	if binary.Read(rd, binary.BigEndian, &blocksize32) != nil {
		return
	}
	blocksize = int(blocksize32)

	hashes = make([]*WeakStrongHash, 0)
	var dobreak bool

	for {
		p := make([]byte, adler32.Size+sha1.Size)

		n, err := rd.Read(p)
		if err != nil {
			if err == io.EOF {
				if n == 0 {
					break
				} else {
					dobreak = true
				}
			} else {
				log.Println("Error reading from sig:", err)
				return
			}
		}

		var weak uint32
		binary.Read(bytes.NewReader(p), binary.BigEndian, &weak)

		hash := &WeakStrongHash{
			Weak:   weak,
			Strong: p[adler32.Size:],
		}
		hashes = append(hashes, hash)

		if dobreak {
			break
		}
	}
	return
}
