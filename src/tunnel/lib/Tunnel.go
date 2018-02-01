package tunnel

import (
	"fmt"
	"io"
	"net"

	"golang.org/x/crypto/ssh"
)

// Credits @svett:
// https://gist.github.com/svett/5d695dcc4cc6ad5dd275

// Credits @josephspurrier
// https://gist.github.com/josephspurrier/e83bcdbf9e6865500004

// Tunnel represents the tunneling
// components: source, proxy and target
// plus the config for the proxy
type Tunnel struct {
	Source *Endpoint
	Proxy  *Endpoint
	Target *Endpoint
	Config *ssh.ClientConfig
}

// Start starts a listener on the Source Server. Once connected it spawns
// a forwarding session (Forward())
func (tunnel *Tunnel) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", tunnel.Source.Port))
	if err != nil {
		fmt.Printf("Could not create Source Server %s\n", err)
		return err
	}

	defer listener.Close()

	return tunnel.StartFromListener(listener)
}

// StartFromConnection starts a forwarding session (Forward()) right
// from an existing Connection
func (tunnel *Tunnel) StartFromConnection(connection net.Conn) error {
	go tunnel.Forward(connection)

	return nil
}

// StartFromListener starts a forwarding session (Forward()) right
// when a connection is established
func (tunnel *Tunnel) StartFromListener(listener net.Listener) error {

	for {
		connection, err := listener.Accept()
		if err != nil {
			return err
		}

		tunnel.StartFromConnection(connection)
	}
}

// Forward connectes to the SSH Server, then connecting
// to the Target Server
func (tunnel *Tunnel) Forward(conn net.Conn) {
	sshconn, err := ssh.Dial("tcp", tunnel.Proxy.String("tcp"), tunnel.Config)
	if err != nil {
		fmt.Printf("Could not connect to SSH-Proxy Server: %s\n", err)
		return
	}

	connection, err := sshconn.Dial("tcp", tunnel.Target.String("tcp"))
	if err != nil {
		fmt.Printf("Could not connect to Target Server %s\n", err)
	}

	copy := func(write, read net.Conn) {
		_, err := io.Copy(write, read)
		if err != nil {
			fmt.Printf("Connection Copy Error: %s\n", err)
			return
		}
	}

	go copy(conn, connection)
	go copy(connection, conn)
}

// Dial connects to the Source Server
func (tunnel *Tunnel) Dial() (net.Conn, error) {
	return net.Dial("tcp", tunnel.Source.String("tcp"))
}
