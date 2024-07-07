package stream

import (
	"compress/gzip"
	"io"

	"gopkg.in/yaml.v3"
)

func gzip_decoder(node *yaml.Node, r io.ReadCloser) (io.ReadCloser, error) {
	check_type(node, "gzip")
	return gzip.NewReader(r)
}

func gzip_encoder(node *yaml.Node, w io.WriteCloser) (io.WriteCloser, error) {
	check_type(node, "gzip")
	return gzip.NewWriterLevel(w, gzip.DefaultCompression)
}

func init() {
	RegisterDecoderStream("gzip", gzip_decoder)
	RegisterEncoderStream("gzip", gzip_encoder)
}
