package stream

import (
	"io"

	"github.com/golang/snappy"
	"gopkg.in/yaml.v3"
)

func snappy_decoder(node *yaml.Node, r io.ReadCloser) (io.ReadCloser, error) {
	check_type(node, "snappy")
	return io.NopCloser(snappy.NewReader(r)), nil
}

func snappy_encoder(node *yaml.Node, w io.WriteCloser) (io.WriteCloser, error) {
	check_type(node, "zlib")

	return snappy.NewBufferedWriter(w), nil
}

func init() {
	RegisterDecoderStream("snappy", snappy_decoder)
	RegisterEncoderStream("snappy", snappy_encoder)
}
