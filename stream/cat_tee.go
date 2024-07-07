package stream

import (
	"errors"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// CAT syntax:
// input:
//   type: cat
//   child:
//     - type: local
// 	  name: xxx
// 	- type: local
// 	  name: yyy
// 	...

type catReader struct {
	child  []io.ReadCloser
	rindex int
}

func (cr *catReader) Read(b []byte) (int, error) {
	if cr.rindex >= len(cr.child) {
		return -1, io.EOF
	}
	r := cr.child[cr.rindex]
	nr, nmax := 0, len(b)

	for nr < nmax {
		n, err := r.Read(b)
		if n > 0 {
			nr += n
			b = b[n:]
		}
		if err != nil {
			if err != io.EOF {
				return nr, err
			}
			// switch to the next reader
			if cr.rindex++; cr.rindex == len(cr.child) {
				return nr, io.EOF
			}
			r = cr.child[cr.rindex]
		}
	}
	return nr, nil
}

func (t *catReader) Close() error {
	var errs []error
	for _, c := range t.child {
		if err := c.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func cat_input(node *yaml.Node) (io.ReadCloser, error) {
	check_type(node, "cat")
	inputs := getList(node, "child")
	if len(inputs) == 0 {
		return nil, fmt.Errorf("no input set in cat")
	}
	child, err := parseInputList(inputs)
	var cr = &catReader{child: child}
	if err != nil {
		if e := cr.Close(); e != nil {
			err = errors.Join(err, e)
		}
	}
	return cr, err
}

// TEE syntax:
// output:
//   type: tee
//   child:
//     - type: local
// 	  name: xxx
// 	- type: stdout
// 	- type: tcp
// 	...

type teeWriter struct {
	child []io.WriteCloser
}

func (t *teeWriter) Write(b []byte) (int, error) {
	var errs []error
	N := len(b)
	nwrite := N
	for _, w := range t.child {
		n, err := w.Write(b)
		if err != nil {
			errs = append(errs, err)
		} else if n != N {
			// do we need to write again
			errs = append(errs, fmt.Errorf("partial write: %v/%v", n, N))
			nwrite = n
		}
	}
	return nwrite, errors.Join(errs...)
}

func (t *teeWriter) Close() error {
	var errs []error
	for _, c := range t.child {
		if err := c.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func tee_output(node *yaml.Node) (io.WriteCloser, error) {
	check_type(node, "tee")
	outputs := getList(node, "child")
	if len(outputs) == 0 {
		return nil, fmt.Errorf("no output set in tee")
	}
	child, err := parseOutputList(outputs)
	var tw = &teeWriter{child: child}
	if err != nil {
		if e := tw.Close(); e != nil {
			err = errors.Join(err, e)
		}
	}
	return tw, err
}

func init() {
	RegisterInputStream("cat", cat_input)
	RegisterOutputStream("tee", tee_output)
}
