package rproxy

import (
	"crypto/sha1"
)

// Could be replaced with go >= 1.2
func sha1sum(p []byte) (sum []byte) {
	sha := sha1.New()
	sha.Write(p)
	return sha.Sum(nil)
}
