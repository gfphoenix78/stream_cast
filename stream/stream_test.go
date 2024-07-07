package stream

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
	"strings"
	"testing"
)

var config1 = `
app:
  name: My Application
  version: 1.0.0

database:
  host: localhost
  port: 3306
  username: root
  password: secret

input:
  type: cat
  child:
    - type: local
      name: /tmp/a
    - type: local
      name: /tmp/b

encoder:
  - type: gzip

#decoder:
#  - type: gzip

output:
  type: tee
  child:
    - type: local
      name: /tmp/w1
    - type: local
      name: /tmp/w2
`

func TestNewInputConfig(t *testing.T) {
	config, err := NewStream(strings.NewReader(config1))
	if err != nil {
		t.Errorf("Parse InputConfig err: %v", err)
	}
	t.Log(*config)

}

// encode: raw(gzip(zlib(bytes)))
func encode(t *testing.T, r io.Reader, w io.WriteCloser) error {
	// file, err := os.OpenFile("/tmp/output.dat",
	// 	os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
	// 	0600)
	// if err != nil {
	// 	t.Errorf("create file: %v", err)
	// }

	w1 := gzip.NewWriter(w)
	w2 := zlib.NewWriter(w1)
	n, err := io.Copy(w2, r)
	if err != nil {
		t.Errorf("encode %v bytes, error: %v", n, err)
	}
	if err = w2.Close(); err != nil {
		panic(err)
	}
	if err = w1.Close(); err != nil {
		panic(err)
	}
	if err = w.Close(); err != nil {
		panic(err)
	}
	return err
}

// decode: zlib(gzip(raw))
func decode(t *testing.T, r io.ReadCloser) (io.ReadCloser, error) {
	r1, err := zlib.NewReader(r)
	if err != nil {
		t.Errorf("new zlib reader: %v", err)
	}
	r2, err := gzip.NewReader(r1)
	if err != nil {
		t.Errorf("new gzip reader: %v", err)
	}

	return r2, err

}

func compare(r1, r2 io.Reader) (bool, error) {
	b1 := make([]byte, 1<<16)
	b2 := make([]byte, 1<<16)
	end := false
	for !end {
		n1, e1 := r1.Read(b1)
		n2, e2 := r2.Read(b2)
		end = e1 == io.EOF && e2 == io.EOF
		if n1 != n2 || (e1 != nil && e1 != io.EOF) || (e2 != nil && e2 != io.EOF) {
			return false, fmt.Errorf("bytes: %v - %v, err: '%v' - '%v'", n1, n2, e1, e2)
		}
		if bytes.Compare(b1[:n1], b2[:n2]) != 0 {
			return false, nil
		}
	}
	return true, nil
}

type nopWriteCloser struct {
	io.Writer
}

func (nop *nopWriteCloser) Close() error {
	if c, ok := nop.Writer.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

func TestEncoderDecoder(t *testing.T) {
	var data = `
	dkfahsiodfjnawjinefijwqbneiurqbwiuerhtniewnbijfbnisjkbdfuiqhw9835y2387y53897y78
	`
	var buffer bytes.Buffer

	err := encode(t, strings.NewReader(data), &nopWriteCloser{&buffer})
	if err != nil {
		t.Errorf("encode err: %v", err)
	}
	rc, err := decode(t, io.NopCloser(&buffer))
	if err != nil {
		t.Errorf("encode err: %v", err)
	}
	defer rc.Close()

	eq, err := compare(strings.NewReader(data), rc)
	if err != nil {
		t.Errorf("compare: %v", err)
	}
	if !eq {
		t.Errorf("Compare not equal")
	}
	t.Log("encode/decode OK")
}
