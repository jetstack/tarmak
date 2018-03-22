// Copyright (c) 2014, Alec Thomas
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
//  - Redistributions of source code must retain the above copyright notice, this
//    list of conditions and the following disclaimer.
//  - Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation
//    and/or other materials provided with the distribution.
//  - Neither the name of SwapOff.org nor the names of its contributors may
//    be used to endorse or promote products derived from this software without
//    specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Package multiplex provides multiplexed streams over a single underlying
// transport `io.ReadWriteCloser`.
//
// Any system that requires a large number of independent TCP connections
// could benefit from this package, by instead having each client maintain a
// single multiplexed connection. There is essentially very little cost to
// creating new channels, or maintaining a large number of open channels.
// Ideal for long term waiting.
//
// An interesting side-effect of this multiplexing is that once the underlying
// connection has been established, each end of the connection can both
// `Accept()` and `Dial()`. This allows for elegant push notifications and
// other interesting approaches.
//
// Documentation
//
// Can be found  on [godoc.org](http://godoc.org/github.com/alecthomas/multiplex) or below.
//
// Example Server
//
//	ln, err := net.Listen("tcp", ":1234")
//	for {
//	    conn, err := ln.Accept()
//	    go func(conn net.Conn) {
//	        mx := multiplex.MultiplexedServer(conn)
//	        for {
//	            c, err := mx.Accept()
//	            go handleConnection(c)
//	        }
//	    }()
//	}
//
// Example Client
//
// Connect to a server with a single TCP connection, then create 10K channels
// over it and write "hello" to each.
//
//	conn, err := net.Dial("tcp", "127.0.0.1:1234")
//	mx := multiplex.MultiplexedClient(conn)
//
//	for i := 0; i < 10000; i++ {
//	    go func() {
//	        c, err := mx.Dial()
//	        n, err := c.Write([]byte("hello"))
//	        c.Close()
//	    }()
//	}
package multiplex

import (
	"encoding/binary"
	"errors"
	"io"
	"sync"
	"sync/atomic"

	"gopkg.in/tomb.v1"
)

// Packet flags.
const (
	SYN = 1 << iota
	RST = 1 << iota
)

const (
	// FragmentSize (in bytes) of packet fragments.
	FragmentSize = 1024
)

var (
	// ErrInvalidChannel is returned when an attempt is made to write to an invalid channel.
	ErrInvalidChannel = errors.New("invalid channel")
)

type packet struct {
	id      uint32
	flags   uint8
	payload []byte
}

type MultiplexedStream struct {
	id       uint32
	conn     io.ReadWriteCloser
	tomb     tomb.Tomb
	channels map[uint32]*Channel
	lock     sync.Mutex
	in       chan *packet
	out      chan *packet
	accept   chan *Channel
}

func newMultiplexer(id uint32, conn io.ReadWriteCloser) *MultiplexedStream {
	m := &MultiplexedStream{
		id:       id,
		conn:     conn,
		channels: make(map[uint32]*Channel),
		in:       make(chan *packet, 1024),
		out:      make(chan *packet, 1024),
		accept:   make(chan *Channel, 64),
	}
	go m.reader()
	go m.run()
	return m
}

// MultiplexedServer creates a new multiplexed server-side stream.
func MultiplexedServer(conn io.ReadWriteCloser) *MultiplexedStream {
	return newMultiplexer(0, conn)
}

// MultiplexedClient creates a new multiplexed client-side stream.
func MultiplexedClient(conn io.ReadWriteCloser) *MultiplexedStream {
	return newMultiplexer(1, conn)
}

// Read packets from the connection and feed them into the in channel.
func (m *MultiplexedStream) reader() {
	var err error

	for m.tomb.Err() == tomb.ErrStillAlive {
		var flags uint8
		var id, size uint32
		if err = binary.Read(m.conn, binary.BigEndian, &id); err != nil {
			break
		}
		if err = binary.Read(m.conn, binary.BigEndian, &size); err != nil {
			break
		}
		flags = uint8(size >> 24)
		size = size & 0xffffff

		payload := make([]byte, size)
		_, err = io.ReadFull(m.conn, payload)
		if err != nil {
			break
		}

		p := &packet{
			id:      id,
			flags:   flags,
			payload: payload,
		}
		m.in <- p
	}

	m.tomb.Kill(err)
}

func (m *MultiplexedStream) run() {
	defer m.tomb.Done()
	var err error

loop:
	for m.tomb.Err() == tomb.ErrStillAlive {
		select {
		// Received packet from peer.
		case p := <-m.in:
			m.lock.Lock()
			ch, ok := m.channels[p.id]
			m.lock.Unlock()

			// No existing channel registered, create a new one.
			// TODO: Handle expired channels correctly.
			if !ok {
				if p.flags&SYN == 0 {
					err = ErrInvalidChannel
					break loop
				}
				ch = newChannel(p.id, m.out, &m.tomb)
				m.lock.Lock()
				m.channels[p.id] = ch
				m.lock.Unlock()

				m.accept <- ch
			}

			// Received a RST, close the channel.
			if p.flags&RST != 0 {
				m.lock.Lock()
				delete(m.channels, p.id)
				m.lock.Unlock()
				ch.Close()
				ch = nil
			}

			if len(p.payload) != 0 {
				if _, err = ch.mw.Write(p.payload); err != nil {
					break loop
				}
			}

		// Send packet from local channel to peer.
		case p := <-m.out:
			if err = binary.Write(m.conn, binary.BigEndian, p.id); err != nil {
				break loop
			}
			packed := (uint32(len(p.payload)) & 0xffffff) | (uint32(p.flags) << 24)
			if err = binary.Write(m.conn, binary.BigEndian, packed); err != nil {
				break loop
			}
			if _, err = m.conn.Write(p.payload); err != nil {
				break loop
			}

			// MultiplexedStream has been killed.
		case <-m.tomb.Dying():
			break loop
		}
	}

	m.tomb.Kill(err)
	m.conn.Close()
}

func (m *MultiplexedStream) Close() error {
	m.tomb.Kill(io.EOF)
	return m.tomb.Wait()
}

func (m *MultiplexedStream) Accept() (*Channel, error) {
	select {
	case ch := <-m.accept:
		return ch, nil
	case <-m.tomb.Dying():
		return nil, m.tomb.Err()
	}
}

// Dial the remote end, creating a new multiplexed channel.
func (m *MultiplexedStream) Dial() (*Channel, error) {
	if err := m.tomb.Err(); err != tomb.ErrStillAlive {
		return nil, err
	}

	id := atomic.AddUint32(&m.id, 2)
	ch := newChannel(id, m.out, &m.tomb)
	ch.out <- &packet{id: ch.id, flags: SYN}

	m.lock.Lock()
	defer m.lock.Unlock()
	m.channels[id] = ch
	return ch, nil
}

// A Channel managed by the multiplexer.
type Channel struct {
	id   uint32
	cr   *io.PipeReader // Channel reads from here (mw).
	mw   *io.PipeWriter // MultiplexedStream writes to here (cr).
	out  chan *packet   // Channel writes packets to here.
	tomb tomb.Tomb
}

func newChannel(id uint32, out chan *packet, tomb *tomb.Tomb) *Channel {
	cr, mw := io.Pipe()
	ch := &Channel{
		id:  id,
		cr:  cr,
		mw:  mw,
		out: out,
	}
	go ch.link(tomb)
	return ch
}

// Link this channel's Tomb to the MultiplexedStream's Tomb.
func (c *Channel) link(tomb *tomb.Tomb) {
	defer c.tomb.Done()

	err := io.EOF

	select {
	case <-tomb.Dying():
		// MultiplexedStream died, not much we can do from here so we just propagate the error.
		err = tomb.Err()

	case <-c.tomb.Dying():
		// MultiplexedStream is still alive (?) send RST packet.
		p := &packet{
			id:    c.id,
			flags: RST,
		}
		c.out <- p

	}

	err = c.maybePipeError(err)
	c.cr.CloseWithError(err)
	c.mw.CloseWithError(err)
}

// Read bytes from a multiplexed channel.
func (c *Channel) Read(b []byte) (int, error) {
	if err := c.tomb.Err(); err != tomb.ErrStillAlive {
		return 0, err
	}
	n, err := c.cr.Read(b)
	return n, c.maybePipeError(err)
}

// Write bytes to a multiplexed channel. The underlying implementation will
// fragment the payload into FragmentSize chunks to prevent starvation of other
// channels.
func (c *Channel) Write(b []byte) (int, error) {
	n := 0

	for i, err := 0, c.tomb.Err(); i < len(b) && err == tomb.ErrStillAlive; i, err = i+FragmentSize, c.tomb.Err() {
		l := len(b) - i
		if l > FragmentSize {
			l = FragmentSize
		}

		p := &packet{id: c.id, payload: b[i:l]}
		c.out <- p
		n += len(p.payload)
	}

	return n, c.maybePipeError(c.tomb.Err())
}

// Don't expose io.ErrClosedPipe.
func (c *Channel) maybePipeError(err error) error {
	switch err {
	case io.ErrClosedPipe:
		err = c.tomb.Err()
		if err == tomb.ErrStillAlive || err == tomb.ErrDying {
			err = io.EOF
		}

	case tomb.ErrStillAlive:
		err = nil

	case tomb.ErrDying:
		err = io.EOF
	}
	return err
}

// Close a multiplexed channel.
func (c *Channel) Close() error {
	c.tomb.Kill(io.EOF)
	// If the channel was terminated due to some other error, return that.
	if err := c.tomb.Wait(); err != io.EOF {
		return err
	}
	return nil
}
