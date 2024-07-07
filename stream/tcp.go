package stream

import (
	"bytes"
	"fmt"
	"io"
	"net"

	"gopkg.in/yaml.v3"
)

type tcp_config struct {
	Type  string
	Host  string
	Port  string
	Role  string // server : client
	Token string // 4 bytes-length + token
}

type conn_rw struct {
	net.Conn
	token []byte // token to compare
	to    []byte // needs be handled
}

type tcpReader struct {
	net.Conn
	token  []byte // token to compare
	buffer []byte
	npos   int
}

func (tr *tcpReader) Read(b []byte) (int, error) {
	ntoken := len(tr.token)
	for tr.npos < ntoken {
		n, err := tr.Conn.Read(tr.buffer[tr.npos:])
		if n > 0 {
			tr.npos += n
		}
		if err != nil {
			return 0, err
		}
		if tr.npos == ntoken && bytes.Compare(tr.buffer, tr.token) != 0 {
			return 0, fmt.Errorf("authorized token doesn't match")
		}
	}

	return tr.Conn.Read(b)
}

func tcp_prepare(config *tcp_config, node *yaml.Node) error {
	err := node.Decode(config)
	if err != nil {
		return err
	}
	if config.Role != "" && config.Role != "server" && config.Role != "client" {
		return fmt.Errorf("invalid role value: %v", config.Role)
	}
	return nil
}

// read/recv bytes, the first is token
func tcp_input(node *yaml.Node) (io.ReadCloser, error) {
	var config tcp_config
	if err := tcp_prepare(&config, node); err != nil {
		return nil, err
	}
	raddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(config.Host, config.Port))
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		return nil, err
	}
	if config.Token == "" {
		// no token
		return conn, nil
	}
	tcpr := &tcpReader{
		Conn:  conn,
		token: []byte(config.Token),
	}
	tcpr.buffer = make([]byte, len(tcpr.token))

	return tcpr, nil
}

type tcpWriter struct {
	net.Conn
	token []byte // token to compare
	npos  int
}

func (tr *tcpWriter) Write(b []byte) (int, error) {
	ntoken := len(tr.token)
	for tr.npos < ntoken {
		n, err := tr.Conn.Write(tr.token[tr.npos:])
		if n > 0 {
			tr.npos += n
		}
		if err != nil {
			return 0, err
		}
	}

	return tr.Conn.Write(b)
}

func tcp_output(node *yaml.Node) (io.WriteCloser, error) {
	var config tcp_config
	if err := tcp_prepare(&config, node); err != nil {
		return nil, err
	}
	raddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(config.Host, config.Port))
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		return nil, err
	}
	if config.Token == "" {
		// no token
		return conn, nil
	}

	tcpr := &tcpWriter{
		Conn:  conn,
		token: []byte(config.Token),
	}
	return tcpr, nil
}

func init() {
	RegisterInputStream("tcp", tcp_input)
	RegisterOutputStream("tcp", tcp_output)
}
