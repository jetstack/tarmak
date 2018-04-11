// Copyright Jetstack Ltd. See LICENSE for details.
package connector

import (
	"bytes"
	"fmt"
	"net"
	"time"
)

const (
	providerSocket = "provider.sock"
	ETX            = byte(3)
	EOT            = byte(4)
)

func (c *Connector) NewServer() error {
	ln, err := net.Listen("unix", providerSocket)
	if err != nil {
		return fmt.Errorf("unable to listen to provider socket: %v", err)
	}

	c.server = ln

	c.log.Infof("Connector server started.")

	return nil
}

func (c *Connector) CloseServer() error {
	if err := c.server.Close(); err != nil {
		return fmt.Errorf("failed to close connector server: %v", err)
	}

	c.log.Infof("Connector server closed.")

	return nil
}

func (c *Connector) StartServer() {
	for {
		select {
		case <-c.stopCh:
			return

		default:
			conn, err := c.AcceptProvider()
			if err != nil {
				continue
			}

			bytes, err := c.HandleConnection(conn)
			if err != nil {
				c.log.Errorf("error handling connection: %v", err)
				continue
			}
			c.log.Debugf("Received from provider: %s\n", bytes)

			reply, err := c.ForwardProviderRequest(bytes)
			if err != nil {
				c.log.Errorf("error forwarding request to rpc server: %v", err)
				continue
			}
			c.log.Debugf("Received from rpc server: %s\n", reply)

			if err := c.SendProvider(conn, reply); err != nil {
				c.log.Errorf("failed to forward response to provider: %v", err)
				continue
			}
			c.log.Infof("Successfully handled provider request")

			if err := conn.Close(); err != nil {
				c.log.Errorf("failed to close connection to provider: %v", err)
			}
		}
	}
}

func (c *Connector) AcceptProvider() (net.Conn, error) {
	conn, err := c.server.Accept()
	if err != nil {
		return nil, fmt.Errorf("unable to accept from provider socket: %v", err)
	}

	return conn, nil
}

func (c *Connector) HandleConnection(conn net.Conn) ([]byte, error) {
	var buff []byte
	b := make([]byte, 1)

	for {

		// Reading may cause a hang, this will timeout the connection
		ch := make(chan struct{})
		go func() {
			ticker := time.NewTicker(time.Second * 10)

			select {
			case <-ch:
				return
			case <-ticker.C:
				ticker.Stop()
				conn.Close()
			}
		}()

		_, err := conn.Read(b)
		if err != nil {
			return nil, fmt.Errorf("failed to read byte from provider: %v", err)
		}

		close(ch)

		if b[0] == EOT {
			break
		}

		buff = append(buff, b...)
	}

	return buff, nil
}

func (c *Connector) ForwardProviderRequest(b []byte) (reply string, err error) {
	f, args, err := c.decodeMessage(b)
	if err != nil {
		return "", fmt.Errorf("failed to decode message from provider: %v", err)
	}

	switch f {
	case "BastionInstanseStatus":
		reply, err = c.CallBastionInstanceStatus(args)
		if err != nil {
			return "", err
		}

	case "VaultClusterStatus":
		reply, err = c.CallVaultClusterStatus(args)
		if err != nil {
			return "", err
		}

	case "VaultInstanceRoleStatus":
		reply, err = c.CallVaultClusterStatus(args)
		if err != nil {
			return "", err
		}

	case "Handshake":
		reply, err = c.CallHandshake()
		if err != nil {
			return "", err
		}

	default:
		return "", fmt.Errorf("RPC function call not supported: %s", f)
	}

	return reply, nil
}

func (c *Connector) decodeMessage(b []byte) (f string, args []string, err error) {
	message := bytes.Split(b, []byte{ETX})

	if len(message) < 1 {
		return "", nil, fmt.Errorf("message malformed or does not contain a function: %s\n", string(b))
	}

	f = string(message[0])
	if len(message) > 1 {
		for _, a := range message[1:] {
			args = append(args, string(a))
		}
	}

	return f, args, nil
}

func (c *Connector) SendProvider(conn net.Conn, message string) error {
	for _, b := range []byte(message) {
		if _, err := conn.Write([]byte{b}); err != nil {
			return fmt.Errorf("error sending byte to connection: %v", err)
		}
	}

	if _, err := conn.Write([]byte{EOT}); err != nil {
		return fmt.Errorf("error sending EOT to connection: %v", err)
	}

	c.log.Debugf("Sent bytes to provider.")

	return nil
}
