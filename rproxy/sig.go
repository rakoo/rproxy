package rproxy

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"math"
)

// Simple, stupid differ: adler-32 as weak, sha1 as strong

const MAX_SIG_SIZE = 1024

var (
	WrongMagic        = errors.New("Wrong magic")
	CantReadVersion   = errors.New("Can't read version")
	WrongVersion      = errors.New("WrongVersion")
	CantReadBlocksize = errors.New("Can't read blocksize")
)

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

		p = p[:n]

		h := &WeakStrongHash{
			Weak:   Checksum(p),
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
	binary.Write(&out, binary.BigEndian, uint32(2)) // version
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

	maxBlocks := math.Ceil(float64(maxSize) / float64(ROLLSUM_SIZE+sha1.Size))
	return uint32(math.Ceil(float64(inputsize) / maxBlocks))
}

func readSig(sig []byte) (blocksize int, hashes []*WeakStrongHash, err error) {

	rd := bytes.NewReader(sig)

	magic := make([]byte, 6)
	rd.Read(magic)
	if string(magic) != "rproxy" {
		log.Println("Wrong file format: magic is not \"rproxy\")")
		err = WrongMagic
		return
	}

	var version uint32
	if binary.Read(rd, binary.BigEndian, &version) != nil {
		err = CantReadVersion
		return
	}

	if version != 2 {
		log.Println("Expected version 2, got", version)
		err = WrongVersion
		return
	}

	var blocksize32 uint32
	if binary.Read(rd, binary.BigEndian, &blocksize32) != nil {
		err = CantReadBlocksize
		return
	}
	blocksize = int(blocksize32)

	hashes = make([]*WeakStrongHash, 0)

	for {
		var weak uint32
		err = binary.Read(rd, binary.BigEndian, &weak)
		if err == io.EOF {
			err = nil
			break
		}

		strong := make([]byte, sha1.Size)
		rd.Read(strong)

		hash := &WeakStrongHash{
			Weak:   weak,
			Strong: strong,
		}
		hashes = append(hashes, hash)
	}
	return
}
