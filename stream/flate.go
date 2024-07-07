package stream

import (
	"compress/flate"
	"io"

	"gopkg.in/yaml.v3"
)

// support options: level, dict

func flate_decoder(node *yaml.Node, r io.ReadCloser) (io.ReadCloser, error) {
	check_type(node, "flate")
	return flate.NewReader(r), nil
}

func flate_encoder(node *yaml.Node, w io.WriteCloser) (io.WriteCloser, error) {
	check_type(node, "flate")
	return flate.NewWriter(w, flate.DefaultCompression)
}

func init() {
	RegisterDecoderStream("flate", flate_decoder)
	RegisterEncoderStream("flate", flate_encoder)
}
