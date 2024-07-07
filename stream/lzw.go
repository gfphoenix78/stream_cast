package stream

import (
	"compress/lzw"
	"io"

	"gopkg.in/yaml.v3"
)

// support options: order, lintwidth

func lzw_decoder(node *yaml.Node, r io.ReadCloser) (io.ReadCloser, error) {
	check_type(node, "lzw")
	return lzw.NewReader(r, lzw.LSB, 8), nil
}

func lzw_encoder(node *yaml.Node, w io.WriteCloser) (io.WriteCloser, error) {
	check_type(node, "lzw")
	return lzw.NewWriter(w, lzw.LSB, 8), nil
}

func init() {
	RegisterDecoderStream("lzw", lzw_decoder)
	RegisterEncoderStream("lzw", lzw_encoder)
}
