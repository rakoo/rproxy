package rproxy

// Use something close to VCDIFF's format
const (
	ADD = iota
	DATA
	CHECKSUM
)

type Op struct {
	// ADD or DATA
	code int

	// Only if code == DATA
	data []byte

	// Only if code == COPY
	index  uint32
	length uint32
}
