package stream

import (
	"compress/zlib"
	"io"

	"gopkg.in/yaml.v3"
)

// support options: level, dict

func zlib_decoder(node *yaml.Node, r io.ReadCloser) (io.ReadCloser, error) {
	check_type(node, "zlib")
	return zlib.NewReader(r)
}

func zlib_encoder(node *yaml.Node, w io.WriteCloser) (io.WriteCloser, error) {
	check_type(node, "zlib")

	return zlib.NewWriterLevel(w, zlib.DefaultCompression)
}

func init() {
	RegisterDecoderStream("zlib", zlib_decoder)
	RegisterEncoderStream("zlib", zlib_encoder)
}
