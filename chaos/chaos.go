package chaos

import (
	"crypto/rand"
	"io"
)

// Writer 不老实的 Writer
type Writer struct {
	// offset     byte
	// complement byte
	Writer io.Writer
}

// WriteChaos 随机写入指定长度的数据
func (w *Writer) WriteChaos(length int) error {
	buf := make([]byte, length)
	r := rand.Reader
	_, err := r.Read(buf)
	if err != nil {
		return err
	}

	_, err = w.Writer.Write(buf)

	return err
}

func (w *Writer) Write(p []byte) (n int, err error) {
	buf := make([]byte, len(p))
	copy(buf, p)

	// var t int
	for i, b := range buf {
		if b >= 0x80 {
			buf[i] = b - 0x80
		} else {
			buf[i] = b + 0x80
		}
	}

	return w.Writer.Write(buf)
}

// Reader 不老实的 Reader
type Reader struct {
	// offset     byte
	// complement byte
	Bytes  []byte
	Offset int
}

func (r *Reader) Read(p []byte) (n int, err error) {
	length := len(r.Bytes)
	remain := length - r.Offset
	if remain <= 0 {
		return 0, io.EOF
	}
	length = len(p)
	if length > remain {
		err = io.EOF
	} else {
		remain = length
	}

	for n = 0; n < remain; n++ {
		b := r.Bytes[r.Offset+n]
		if b >= 0x80 {
			p[n] = b - 0x80
		} else {
			p[n] = b + 0x80
		}
	}
	r.Offset += remain

	return
}
