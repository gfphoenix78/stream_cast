package stream

import (
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type localInputFile struct {
	Type string
	Name string
}
type localOutputFile struct {
	Type string
	Name string
	// perm
	// flag
}

func local_input(input *yaml.Node) (io.ReadCloser, error) {
	var config localInputFile
	err := input.Decode(&config)
	if err != nil {
		return nil, err
	}
	// assert config.Type = "local"
	return os.Open(config.Name)
}

func local_output(node *yaml.Node) (io.WriteCloser, error) {
	var config localOutputFile
	err := node.Decode(&config)
	if err != nil {
		return nil, err
	}
	// assert config.Type = "local"
	return os.OpenFile(config.Name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
}

func init() {
	RegisterInputStream("local", local_input)
	RegisterOutputStream("local", local_output)
}
