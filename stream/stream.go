package stream

import (
	"errors"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// input:
//
//	type: stdin/local/tcp/unix/http/https/...
//	...
//
// decoder:
//   - type: gzip/zlib/...
//     level: 12
//   - type: gzip/zlib/...
//     options: ...
//
// encoder:
//   - type: gzip/zlib/...
//     level: 12
//   - type: gzip/zlib/...
//     options: ...
//
// output:
//
//	type: stdout/local/tcp/unix/http/https/...
//	...

// input: () -> io.ReadCloser
// encoder: (io.ReadCloser) io.ReadCloser
// decoder: (io.WriteCloser) io.WriteCloser
// output: () -> io.WriteCloser
//
// io.Copy(w, r)

type InputFunc func(node *yaml.Node) (io.ReadCloser, error)
type OutputFunc func(node *yaml.Node) (io.WriteCloser, error)
type DecoderFunc func(node *yaml.Node, r io.ReadCloser) (io.ReadCloser, error)
type EncoderFunc func(node *yaml.Node, w io.WriteCloser) (io.WriteCloser, error)

var input_funcs = make(map[string]InputFunc)
var output_funcs = make(map[string]OutputFunc)
var decoder_funcs = make(map[string]DecoderFunc)
var encoder_funcs = make(map[string]EncoderFunc)

func RegisterInputStream(name string, fn InputFunc) {
	input_funcs[name] = fn
}
func RegisterOutputStream(name string, fn OutputFunc) {
	output_funcs[name] = fn
}
func RegisterDecoderStream(name string, fn DecoderFunc) {
	decoder_funcs[name] = fn
}
func RegisterEncoderStream(name string, fn EncoderFunc) {
	encoder_funcs[name] = fn
}

type Stream struct {
	input   io.ReadCloser
	decoder []io.ReadCloser
	encoder []io.WriteCloser
	output  io.WriteCloser

	closed bool
}

func getChildByTag(node *yaml.Node, name string) int {
	for i, n := 0, len(node.Content); i < n; i += 2 {
		if node.Content[i].Value == name {
			return i
		}
	}
	return -1
}
func getInputOuputMap(node *yaml.Node, name string) (*yaml.Node, error) {
	i := getChildByTag(node, name)
	if i < 0 || i+1 >= len(node.Content) {
		return nil, fmt.Errorf("'%v' not found", name)
	}
	if node.Content[i].Kind != yaml.ScalarNode ||
		node.Content[i+1].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("invalid '%v', needs a map", name)
	}
	return node.Content[i+1], nil
}

func getList(node *yaml.Node, name string) []*yaml.Node {
	i := getChildByTag(node, name)
	if i < 0 || i+1 >= len(node.Content) {
		return nil
	}
	if node.Content[i].Kind != yaml.ScalarNode ||
		node.Content[i+1].Kind != yaml.SequenceNode {
		return nil
	}
	return node.Content[i+1].Content
}

func getElementType(node *yaml.Node) (string, error) {
	if node.Kind != yaml.MappingNode {
		return "", fmt.Errorf("unexpected node kind: %v", node.Kind)
	}
	i := getChildByTag(node, "type")
	if i < 0 || i+1 >= len(node.Content) {
		return "", fmt.Errorf("`input`. type is missing ")
	}
	return node.Content[i+1].Value, nil
}

func parseInput(node *yaml.Node) (io.ReadCloser, error) {
	// input
	input, err := getInputOuputMap(node, "input")
	if err != nil {
		return nil, err
	}
	name, err := getElementType(input)
	if err != nil {
		return nil, fmt.Errorf("no `type` found in input stream")
	}
	fn, ok := input_funcs[name]
	if !ok {
		return nil, fmt.Errorf("no input registry: %v", name)
	}
	return fn(input)
}

func parseOutput(node *yaml.Node) (io.WriteCloser, error) {
	// output
	output, err := getInputOuputMap(node, "output")
	if err != nil {
		return nil, err
	}
	name, err := getElementType(output)
	if err != nil {
		return nil, fmt.Errorf("no `type` found in output stream")
	}
	fn, ok := output_funcs[name]
	if !ok {
		return nil, fmt.Errorf("no input registry: %v", name)
	}
	return fn(output)
}

func parseInputList(nodes []*yaml.Node) ([]io.ReadCloser, error) {
	var list []io.ReadCloser
	var rc io.ReadCloser
	for _, node := range nodes {
		name, err := getElementType(node)
		if err != nil {
			return list, fmt.Errorf("no `type` found in output stream")
		}
		fn, ok := input_funcs[name]
		if !ok {
			return list, fmt.Errorf("no input registry: %v", name)
		}
		if rc, err = fn(node); err != nil {
			return list, err
		}
		list = append(list, rc)
	}
	return list, nil
}

func parseOutputList(nodes []*yaml.Node) ([]io.WriteCloser, error) {
	var list []io.WriteCloser
	var wc io.WriteCloser
	for _, node := range nodes {
		name, err := getElementType(node)
		if err != nil {
			return list, fmt.Errorf("no `type` found in output stream")
		}
		fn, ok := output_funcs[name]
		if !ok {
			return list, fmt.Errorf("no input registry: %v", name)
		}
		if wc, err = fn(node); err != nil {
			return list, err
		}
		list = append(list, wc)
	}
	return list, nil
}

func parseEncoder(list []*yaml.Node, wc io.WriteCloser) ([]io.WriteCloser, error) {
	var encoders []io.WriteCloser
	for _, node := range list {
		if node.Kind != yaml.MappingNode {
			return encoders, fmt.Errorf("invalid yaml format: codec should be map: %v", node)
		}
		typname, err := getElementType(node)
		if err != nil {
			return encoders, err
		}
		fn, ok := encoder_funcs[typname]
		if !ok {
			return encoders, fmt.Errorf("encoder '%v' not found", typname)
		}
		wc, err = fn(node, wc)
		if err != nil {
			return encoders, err
		}
		encoders = append(encoders, wc)
	}
	return encoders, nil
}

func parseDecoder(list []*yaml.Node, rc io.ReadCloser) ([]io.ReadCloser, error) {
	var decoders []io.ReadCloser
	for _, node := range list {
		if node.Kind != yaml.MappingNode {
			return decoders, fmt.Errorf("invalid yaml format: codec should be map: %v", node)
		}
		typname, err := getElementType(node)
		if err != nil {
			return decoders, err
		}
		fn, ok := decoder_funcs[typname]
		if !ok {
			return decoders, fmt.Errorf("decoder '%v' not found", typname)
		}
		rc, err = fn(node, rc)
		if err != nil {
			return decoders, err
		}
		decoders = append(decoders, rc)
	}
	return decoders, nil
}

func (s *Stream) parseEncoder(list []*yaml.Node) error {
	var err error
	s.encoder, err = parseEncoder(list, s.output)
	if err != nil {
		return err
	}
	return err
}

func (s *Stream) parseDecoder(list []*yaml.Node) error {
	var err error
	s.decoder, err = parseDecoder(list, s.input)
	if err != nil {
		return err
	}
	return err
}

// may read bytes that will be blocked
func NewStream(r io.Reader) (*Stream, error) {
	var node yaml.Node
	var stream Stream
	err := yaml.NewDecoder(r).Decode(&node)
	if err != nil {
		return nil, err
	}

	// input
	stream.input, err = parseInput(node.Content[0])
	if err != nil {
		return nil, err
	}

	// output
	stream.output, err = parseOutput(node.Content[0])
	if err != nil {
		return nil, err
	}

	// decoder
	decoder := getList(node.Content[0], "decoder")
	if decoder != nil {
		if err = stream.parseDecoder(decoder); err != nil {
			return nil, err
		}
	}

	// encoder
	encoder := getList(node.Content[0], "encoder")
	if encoder != nil {
		if err = stream.parseEncoder(encoder); err != nil {
			return nil, err
		}
	}

	return &stream, nil
}

func (s *Stream) Reader() io.ReadCloser {
	if n := len(s.decoder); n > 0 {
		return s.decoder[n-1]
	}
	return s.input
}
func (s *Stream) Writer() io.WriteCloser {
	if n := len(s.encoder); n > 0 {
		return s.encoder[n-1]
	}
	return s.output
}
func (s *Stream) Copy() (int64, error) {
	r := s.Reader()
	w := s.Writer()
	return io.Copy(w, r)
}

func (s *Stream) Close() error {
	var errs []error
	if s.closed {
		return nil
	}
	// close reader first
	for i := len(s.decoder) - 1; i >= 0; i-- {
		if e := s.decoder[i].Close(); e != nil {
			errs = append(errs, e)
		}
	}
	if e := s.input.Close(); e != nil {
		errs = append(errs, e)
	}

	// close writer, outer to inner
	for i := len(s.encoder) - 1; i >= 0; i-- {
		if e := s.encoder[i].Close(); e != nil {
			errs = append(errs, e)
		}
	}
	if e := s.output.Close(); e != nil {
		errs = append(errs, e)
	}
	s.closed = true
	return errors.Join(errs...)
}
