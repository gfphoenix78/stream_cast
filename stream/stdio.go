package stream

import (
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

func lookupStdin(node *yaml.Node) (io.ReadCloser, error) {
	check_type(node, "stdin")
	return os.Stdin, nil
}
func lookupStdout(node *yaml.Node) (io.WriteCloser, error) {
	check_type(node, "stdout")
	return os.Stdout, nil
}

func init() {
	RegisterInputStream("stdin", lookupStdin)
	RegisterOutputStream("stdout", lookupStdout)
}
