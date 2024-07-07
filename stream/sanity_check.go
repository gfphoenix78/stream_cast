package stream

import "gopkg.in/yaml.v3"

func check_type(node *yaml.Node, typname string) {

	name, err := getElementType(node)
	if err != nil {
		panic(err)
	}
	if name != typname {
		panic("bug: dispatch to " + typname)
	}
}
