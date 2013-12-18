package rproxy

import (
	"bytes"
	"io"
	"log"
)

func Patch(delta []byte, baseFile io.ReaderAt) (newFile io.Reader, checksum []byte, err error) {
	ops, checksum, err := readDelta(delta)
	if err != nil {
		log.Println("Couldn't patch:", err)
		return
	}

	var wr bytes.Buffer

	for _, op := range ops {
		switch op.code {
		case ADD:
			b := make([]byte, op.length)
			baseFile.ReadAt(b, int64(op.index))
			wr.Write(b)
			break
		case DATA:
			wr.Write(op.data)
			break
		}
	}

	newFile = &wr
	return
}
