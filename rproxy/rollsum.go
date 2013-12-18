/*
Copyright 2011 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Reimplementation of camlistore's rollsum, along with a method to
// calculate the checksum of a whole []byte, and a custom window size
package rproxy

const charOffset = 31

const ROLLSUM_SIZE = 4

type RollSum struct {
	s1, s2     uint32
	windowSize int
}

func NewRollsum(windowSize int) *RollSum {
	return &RollSum{
		s1:         uint32(windowSize * charOffset),
		s2:         uint32(windowSize * (windowSize - 1) * charOffset),
		windowSize: windowSize,
	}
}

func (rs *RollSum) Roll(drop, add uint8) {
	rs.s1 += uint32(add) - uint32(drop)
	rs.s2 += rs.s1 - uint32(rs.windowSize)*uint32(drop+charOffset)
}

func (rs *RollSum) Digest() uint32 {
	return (rs.s1 << 16) | (rs.s2 & 0xffff)
}

func Checksum(p []byte) uint32 {
	s1 := uint32(len(p) * charOffset)
	s2 := uint32(len(p) * (len(p) - 1) * charOffset)

	for _, x := range p {
		s1 += uint32(x)
		s2 += s1 - uint32(len(p))*uint32(charOffset)
	}

	return (s1 << 16) | (s2 & 0xffff)
}
