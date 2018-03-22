// Copyright Jetstack Ltd. See LICENSE for details.
package connector

import (
	"io"
	"net"
	"os"
	"sync"

	"github.com/alecthomas/multiplex"
	"github.com/sirupsen/logrus"

	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

type Proxy struct {
	SocketPath string
	Done       chan struct{}
	log        *logrus.Entry
	listener   net.Listener

	Writer io.Writer
	Logger io.Writer
	Reader io.Reader
}

func NewProxy(socketPath string) *Proxy {

	p := &Proxy{
		SocketPath: socketPath,
		Done:       make(chan struct{}),
		Writer:     os.Stdout,
		Logger:     os.Stderr,
		Reader:     os.Stdin,
	}

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.Out = p.Logger
	p.log = logger.WithFields(logrus.Fields{
		"socket": socketPath,
	})

	return p
}

func (p *Proxy) Start() error {
	p.log.Infoln("starting connector")

	// create the socket to listen on:
	var err error
	p.listener, err = net.Listen("unix", p.SocketPath)
	if err != nil {
		return err
	}

	// handle signals
	signalCh := utils.BasicSignalHandler(p.log)
	go func() {
		// wait for a signal
		<-signalCh
		p.Stop()
	}()

	go p.run()
	return nil
}

func (p *Proxy) Stop() {
	p.log.Infoln("stopping proxy")
	if p.Done != nil {
		close(p.Done)
	}
	if p.listener != nil {
		p.listener.Close()
	}
	p.Done = nil
}

func (p *Proxy) run() {
	closer, _ := utils.NewCloser()
	mx := multiplex.MultiplexedClient(struct {
		io.Reader
		io.Writer
		io.Closer
	}{p.Reader, p.Writer, closer})

	for {
		select {
		case <-p.Done:
			return
		default:
			clientConn, err := p.listener.Accept()
			if err == nil {
				tarmakConn, err := mx.Dial()
				if err != nil {
					p.log.WithField("err", err).Errorln("error connecting to tarmak")
				}
				go p.handle(tarmakConn, clientConn)
			} else {
				p.log.WithField("err", err).Errorln("error accepting connection")
			}
		}
	}
}

func (p *Proxy) handle(tarmak *multiplex.Channel, client net.Conn) {
	p.log.Debugf("handling new unix socket connection: %x", client)
	defer p.log.Debugf("done handling: %x", client)
	defer tarmak.Close()
	defer client.Close()
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go p.copy(tarmak, client, wg)
	go p.copy(client, tarmak, wg)
	wg.Wait()
}

func (p *Proxy) copy(to io.Writer, from io.Reader, wg *sync.WaitGroup) {
	defer wg.Done()
	select {
	case <-p.Done:
		return
	default:
		if _, err := io.Copy(to, from); err != nil {
			p.log.WithField("err", err).Errorln("Error from copy")
			p.Stop()
			return
		}
	}
}
