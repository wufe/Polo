package utils

import (
	"fmt"
	"io"
	"net"
	"sync"

	udp_forward "github.com/1lann/udp-forward"
)

type Forwarder interface {
	Close()
	OnConnect(func(addr string))
	OnDisconnect(func(addr string))
}

func ForwardTCP(source string, dest string) (Forwarder, error) {
	return newTcpForwarder(source, dest)
}

func ForwardUDP(source string, dest string) (Forwarder, error) {
	return udp_forward.Forward(
		source,
		dest,
		udp_forward.DefaultTimeout,
	)
}

type tcpForwarder struct {
	listener           net.Listener
	connectCallback    func(addr string)
	disconnectCallback func(addr string)
	connectionsMutex   struct{ sync.RWMutex }
	connections        map[net.Conn]net.Conn
}

func newTcpForwarder(source string, dest string) (*tcpForwarder, error) {
	f := &tcpForwarder{
		connections: make(map[net.Conn]net.Conn),
	}
	listener, err := net.Listen("tcp4", source)
	if err != nil {
		return nil, fmt.Errorf("error listening: %w", err)
	}

	f.listener = listener

	go func() {
		defer listener.Close()
		for {
			downstreamConn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(downstreamConn net.Conn) {
				defer func() {
					downstreamConn.Close()
					if err := recover(); err != nil {
						fmt.Println("unhandled exception:", err)
					}
				}()
				upstreamConn, err := net.Dial("tcp4", dest) // dest port
				if err != nil {
					// TODO: How to print error?
					return
				}
				if f.connectCallback != nil {
					f.connectCallback(downstreamConn.RemoteAddr().String())
				}

				f.connectionsMutex.Lock()
				f.connections[downstreamConn] = upstreamConn
				f.connectionsMutex.Unlock()

				defer upstreamConn.Close()
				go io.Copy(upstreamConn, downstreamConn)
				io.Copy(downstreamConn, upstreamConn)

				if f.disconnectCallback != nil {
					f.disconnectCallback(downstreamConn.RemoteAddr().String())
				}
				f.connectionsMutex.Lock()
				delete(f.connections, downstreamConn)
				f.connectionsMutex.Unlock()
			}(downstreamConn)
		}
	}()

	return f, nil
}

func (f *tcpForwarder) Close() {
	f.connectionsMutex.RLock()
	defer f.connectionsMutex.RUnlock()

	for downstream, upstream := range f.connections {
		upstream.Close()
		downstream.Close()
	}

	f.listener.Close()
}

func (f *tcpForwarder) OnConnect(callback func(addr string)) {
	f.connectCallback = callback
}

func (f *tcpForwarder) OnDisconnect(callback func(addr string)) {
	f.disconnectCallback = callback
}
