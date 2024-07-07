package stream

import (
	"compress/gzip"
	"io"

	"gopkg.in/yaml.v3"
)

// type gzip_config struct {
// 	Type  string
// 	Level int
// }

func gzip_decoder(node *yaml.Node, r io.ReadCloser) (io.ReadCloser, error) {
	check_type(node, "gzip")
	return gzip.NewReader(r)
}

func gzip_encoder(node *yaml.Node, w io.WriteCloser) (io.WriteCloser, error) {
	check_type(node, "gzip")
	// var config gzip_config
	// err := node.Decode(&config)
	// if err != nil {
	// 	return nil, err
	// }
	// if config.Level != gzip.NoCompression {
	// 	config.Level = gz
	// }
	return gzip.NewWriterLevel(w, gzip.DefaultCompression)
}

func init() {
	RegisterDecoderStream("gzip", gzip_decoder)
	RegisterEncoderStream("gzip", gzip_encoder)
}
