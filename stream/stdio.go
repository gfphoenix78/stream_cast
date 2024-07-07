package stream

import (
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

func std_input(node *yaml.Node) (io.ReadCloser, error) {
	check_type(node, "stdin")
	return os.Stdin, nil
}
func std_output(node *yaml.Node) (io.WriteCloser, error) {
	check_type(node, "stdout")
	return os.Stdout, nil
}

func init() {
	RegisterInputStream("stdin", std_input)
	RegisterOutputStream("stdout", std_output)
}
