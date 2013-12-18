package main

import (
  "bytes"
  "crypto/sha1"
  "github.com/rakoo/rproxy/rproxy"
  "io"
  "log"
  "os"
)

type duplicatingFile struct {
  file *os.File
  raw []byte
}
func NewDuplicatingFile(f *os.File) *duplicatingFile {
  return &duplicatingFile{f, make([]byte, 0)}
}

func (df *duplicatingFile) Read(p []byte) (n int, err error) {
  n, err = df.file.Read(p)
  if err != nil {
    df.raw = append(df.raw, p...)
  }
  return
}

func main() {
  from, size := openOrDie("Obama")
  defer from.Close()

  rawSig := rproxy.Signature(from, size)
  log.Printf("Rawsig is %d bytes long", len(rawSig))


  to, _ := openOrDie("Obama2")
  defer to.Close()
  rawDelta, err := rproxy.Delta(rawSig, to)
  if err != nil {
    log.Fatal(err)
  }
  log.Printf("Delta is %d bytes long", len(rawDelta))

  dest, checksum, err := rproxy.Patch(rawDelta, from)
  if err != nil {
    log.Fatal(err)
  }
  var destBuff bytes.Buffer
  _, err = io.Copy(&destBuff, dest)
  if err != nil {
    log.Fatal(err)
  }

  sha := sha1.New()
  sha.Write(destBuff.Bytes())
  actual := sha.Sum(nil)

  if bytes.Compare(actual, checksum) != 0 {
    log.Printf("Transmission error, invalidated by checksum; Expected %x, got %x", checksum, actual)
  }

  to.Seek(0, os.SEEK_SET)
  expected := sha1sum(to)
  if bytes.Compare(expected, actual) != 0 {
    log.Printf("Different checksums: %x, %x", expected, actual)
  }

  tmp, _ := os.Create("tmp")
  tmp.Write(destBuff.Bytes())
  tmp.Close()

  /*
  dest, err := os.Create("tmp")
  if err != nil {
    log.Panic("Couldn't create tmp file:", err)
  }

  var destBuff bytes.Buffer
  */

}

func openOrDie(filename string) (*os.File, uint64) {
  f, err := os.Open(filename)
  if err != nil {
    log.Fatal("Couldn't open Obama file")
  }

  stat, err := f.Stat()
  if err != nil {
    log.Fatal(err)
  }

  return f, uint64(stat.Size())
}

func sha1sum(in io.Reader) []byte {
  sha := sha1.New()
  io.Copy(sha, in)
  return sha.Sum(nil)
}
