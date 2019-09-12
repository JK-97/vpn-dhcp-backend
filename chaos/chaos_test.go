package chaos

import (
	"bytes"
	"testing"
)

func TestChaosWrite(t *testing.T) {

	buffer := bytes.NewBuffer(make([]byte, 0, 1024))

	w := Writer{
		Writer: buffer,
	}
	msg := "Hello World"
	prefix := 512
	suffix := 128
	w.WriteChaos(prefix)
	w.Write([]byte(msg))
	w.WriteChaos(suffix)

	b := buffer.Bytes()

	r := Reader{
		Bytes:  b,
		Offset: prefix,
	}

	content := make([]byte, len(b)-prefix-suffix)
	_, err := r.Read(content)

	if err != nil {
		t.Error(err)
	}

	if string(content) != msg {
		t.Error(string(content), "!=", msg)
	}
}
