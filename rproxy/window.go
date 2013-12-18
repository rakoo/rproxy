package rproxy

type Window struct {
	Length int
	Filled int // assumes we don't use this for more than 2^32 bytes

	array []byte
	ofs   int
}

func NewWindow(length int) *Window {
	return &Window{
		Length: length,
		array:  make([]byte, length),
	}
}

func (w *Window) Append(b byte) (drop byte) {
	drop = w.array[w.ofs]
	w.array[w.ofs] = b
	w.ofs = (w.ofs + 1) % w.Length

	w.Filled++
	return
}

func (w *Window) Bytes() (p []byte) {
	if w.Filled < w.Length {
		return w.array[:w.ofs]
	}

	p = make([]byte, w.Length)
	copy(p, w.array[w.ofs:])
	copy(p[w.Length-w.ofs:], w.array[:w.ofs])

	return
}

func (w *Window) Reset() {
	w.array = make([]byte, w.Length)
	w.ofs = 0
	w.Filled = 0
}
