//go:build stream_cast
// +build stream_cast

package main

import (
	"flag"
	"log"
	"os"

	stream "github.com/gfphoenix78/stream_cast/v2"
)

var yaml_file string

func main() {
	flag.StringVar(&yaml_file, "config", "", "yaml file to desc")
	flag.Parse()

	file, err := os.Open(yaml_file)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	s, err := stream.NewStream(file)
	if err != nil {
		log.Printf("Parse config file: %v", err)
		os.Exit(1)
	}
	if err = file.Close(); err != nil {
		log.Printf("Close %v: %v", yaml_file, err)
	}

	if n, err := s.Copy(); err != nil {
		log.Println("Copy: %v %v", n, err)
		os.Exit(1)
	}
	if err = s.Close(); err != nil {
		log.Printf("Close: %v", err)
		os.Exit(1)
	}
}
