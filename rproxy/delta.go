package rproxy

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"io"
	"log"
)

var (
	InvalidDeltaMagic   = errors.New("Wrong magic")
	InvalidDeltaVersion = errors.New("WrongVersion")
)

func Delta(sig []byte, newFile io.Reader) (delta []byte, err error) {
	blocksize, hashes, err := readSig(sig)
	if err != nil {
		log.Println("Error when reading sig:", err)
		return
	}

	linear := make(map[string]uint32, len(hashes))
	for i, hash := range hashes {
		linear[string(hash.Strong)] = uint32(i * blocksize)
	}

	hmap := make(map[uint32][]*WeakStrongHash, 0)

	for _, hash := range hashes {
		if hmap[hash.Weak] == nil {
			hmap[hash.Weak] = make([]*WeakStrongHash, 0)
		}
		hmap[hash.Weak] = append(hmap[hash.Weak], hash)
	}

	window := NewWindow(blocksize)
	roller := NewRollsum(blocksize)

	var running bytes.Buffer
	fr := bufio.NewReader(newFile)
	var dobreak bool

	sha := sha1.New()

	var pos uint64
	ops := make([]*Op, 0)

	for {
		c, err := fr.ReadByte()
		if err != nil {
			if err == io.EOF {
				dobreak = true
			} else {
				log.Println("Couldn't read from newfile:", err)
				return nil, err
			}
		}

		if dobreak {

			// Much code just to say that we want weakhash to be equals to
			// Checksum(window.Bytes()). Here we reset roller to a custom
			// size (not blocksize), so that calling Digest() will do the
			// right thing.
			roller = NewRollsum(window.Filled)
			if window.Filled < blocksize {
				for _, x := range window.Bytes() {
					roller.Roll(0, x)
				}
			}

		} else {

			pos++
			sha.Write([]byte{c})
			running.WriteByte(c)
			drop := window.Append(c)

			if window.Filled < blocksize {
				continue
			} else if window.Filled > blocksize {
				roller.Roll(drop, c)
			} else {
				for _, x := range window.Bytes() {
					roller.Roll(0, x)
				}
			}
		}

		weakhash := roller.Digest()

		if _, ok := hmap[weakhash]; !ok {
			if dobreak {
				break
			}
			continue
		}

		stronghash := sha1sum(window.Bytes())
		for _, match := range hmap[weakhash] {
			if bytes.Compare(match.Strong, stronghash) != 0 {
				continue
			}

			if window.Filled > blocksize {
				// There are some new bytes
				data := make([]byte, window.Filled-blocksize)
				copy(data, running.Bytes())
				dataOp := &Op{
					code: DATA,
					data: data,
				}
				ops = append(ops, dataOp)
			}

			// The data we need to ADD is the data that was matched, ie on
			// blocksize bytes. However, for the last block, we may not have
			// filled a whole blocksize of data, so we just take what's
			// there ie window.Filled
			addLength := blocksize
			if dobreak {
				addLength = window.Filled
			}

			addOp := &Op{
				code:   ADD,
				index:  linear[string(stronghash)],
				length: uint32(addLength),
			}
			ops = append(ops, addOp)

			window.Reset()
			roller = NewRollsum(blocksize)
			running.Reset()

			break
		}

		if dobreak {
			break
		}
	}

	if window.Filled > 0 {
		// End of file: flush last block
		ops = append(ops, &Op{
			code: DATA,
			data: running.Bytes(),
		})
	}

	rawOps := serializeDelta(flatten(ops))

	rawOps = append(rawOps, []byte{uint8(CHECKSUM)}...)
	rawOps = append(rawOps, sha.Sum(nil)...)

	return rawOps, nil
}

func findIndex(stronghash []byte, hashes []*WeakStrongHash) uint32 {
	for i, candidate := range hashes {
		if string(candidate.Strong) == string(stronghash) {
			return uint32(i)
		}
	}

	return uint32(0)
}

func serializeDelta(ops []*Op) (raw []byte) {
	var rawBuf bytes.Buffer

	rawBuf.WriteString("rproxy-delta")
	binary.Write(&rawBuf, binary.BigEndian, uint32(1)) // version

	for _, op := range ops {
		binary.Write(&rawBuf, binary.BigEndian, uint8(op.code))
		switch op.code {
		case ADD:
			binary.Write(&rawBuf, binary.BigEndian, op.index)
			binary.Write(&rawBuf, binary.BigEndian, op.length)
		case DATA:
			binary.Write(&rawBuf, binary.BigEndian, uint32(len(op.data)))
			rawBuf.Write(op.data)
		}
	}

	return rawBuf.Bytes()
}

// factorize consecutive ADD ops
func flatten(ops []*Op) (flat []*Op) {
	flat = make([]*Op, 0)

	lastop := ops[0]

	for i := 1; i < len(ops); i++ {
		op := ops[i]

		switch lastop.code {
		case ADD:
			switch op.code {
			case ADD:
				lastop.length += op.length
			case DATA:
				flat = append(flat, lastop)
				lastop = op
			}

		case DATA:
			switch op.code {
			case ADD:
				flat = append(flat, lastop)
				lastop = op
			case DATA: // weird, shouldn't happen...
				lastop.data = append(lastop.data, op.data...)
			}
		}
	}

	flat = append(flat, lastop)

	return
}

func readDelta(raw []byte) (ops []*Op, checksum []byte, err error) {
	if string(raw[:12]) != "rproxy-delta" {
		err = InvalidDeltaMagic
		return
	}

	var version uint32
	binary.Read(bytes.NewReader(raw[12:16]), binary.BigEndian, &version)
	if version != 1 {
		err = InvalidDeltaVersion
		return
	}

	ops = make([]*Op, 0)
	rd := bytes.NewReader(raw[16:])

ops:
	for {
		c, err := rd.ReadByte()
		if err != nil && err == io.EOF {
			break
		}

		switch uint8(c) {
		case ADD:
			newOp := &Op{code: ADD}
			binary.Read(rd, binary.BigEndian, &newOp.index)
			binary.Read(rd, binary.BigEndian, &newOp.length)

			ops = append(ops, newOp)
			break

		case DATA:
			var length uint32
			binary.Read(rd, binary.BigEndian, &length)

			data := make([]byte, length)
			rd.Read(data)

			ops = append(ops, &Op{code: DATA, data: data})
			break
		case CHECKSUM:
			break ops
		}
	}

	var cs bytes.Buffer
	io.Copy(&cs, rd)
	checksum = cs.Bytes()
	return
}
