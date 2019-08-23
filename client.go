// Copyright (c) 2014-2017 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package bitmarkdClient

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"sync"
	"sync/atomic"
	"time"
)

var ErrRPCRequestTimeout = fmt.Errorf("rpc requests timed out")
var ErrRPCConnectionClosed = fmt.Errorf("rpc connection closed")
var ErrCannotEstablishConnection = fmt.Errorf("connection can not be established")

// RPCEmptyArguments is an empty argument for rpc requests
type RPCEmptyArguments struct{}

// PersistentRPCClient is client that will maintain a long-lived connection for requests
type PersistentRPCClient struct {
	sync.Mutex
	*rpc.Client
	address   string
	closed    chan struct{}
	connected bool
	timeout   time.Duration
}

func NewPersistentRPCClient(address string, timeout time.Duration) *PersistentRPCClient {
	return &PersistentRPCClient{
		address: address,
		timeout: timeout,
		closed:  make(chan struct{}),
	}
}

// Call is to make an RPC request. It will whether the connection is still alived.
// If not, it will try to create one.
func (c *PersistentRPCClient) Call(serviceMethod string, args interface{}, reply interface{}) error {
	if !c.connected {
		if err := c.connect(); err != nil {
			return ErrCannotEstablishConnection
		}
	}

	if !c.connected {
		return ErrRPCConnectionClosed
	}

	select {
	case call := <-c.Client.Go(serviceMethod, args, reply, make(chan *rpc.Call, 1)).Done:
		return call.Error
	case <-c.closed:
		return ErrRPCConnectionClosed
	case <-time.After(c.timeout):
		return ErrRPCRequestTimeout
	}
}

// Close will close the current connection
func (c *PersistentRPCClient) Close() error {
	c.Lock()
	defer c.Unlock()

	if !c.connected {
		return nil
	}

	c.connected = false
	close(c.closed)
	return c.Client.Close()
}

// connect will establish a TCP connection to the remote
func (c *PersistentRPCClient) connect() error {
	c.Lock()
	defer c.Unlock()

	if c.connected {
		return nil
	}

	conn, err := tls.Dial("tcp", c.address, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return fmt.Errorf("can not dial to: %s, error: %s", c.address, err.Error())
	}
	c.connected = true
	c.closed = make(chan struct{})
	c.Client = jsonrpc.NewClient(conn)
	return nil
}

// BitmarkdRPCClient is a client to make bitmarkd RPC requests. It maintains
// a list of PersistentRPCClient that will create connections to bitmarkd.
type BitmarkdRPCClient struct {
	sync.RWMutex
	counter   uint32
	clients   map[string]*PersistentRPCClient
	addresses []string
}

// NewBitmarkdRPCClient is to create a rpc client for bitmarkd
func New(addresses []string, timeout time.Duration) *BitmarkdRPCClient {
	clients := map[string]*PersistentRPCClient{}

	for _, addr := range addresses {
		clients[addr] = NewPersistentRPCClient(addr, timeout)
	}

	client := &BitmarkdRPCClient{
		addresses: addresses,
		clients:   clients,
	}

	return client
}

// Close will terminate all rpc clients
func (bc *BitmarkdRPCClient) Close() {
	bc.Lock()
	defer bc.Unlock()
	for _, c := range bc.clients {
		c.Close()
	}
}

// client will return a client based on the addresses list in a round-robin manner
func (bc *BitmarkdRPCClient) client() *PersistentRPCClient {
	bc.RLock()
	defer bc.RUnlock()
	counter := atomic.AddUint32(&bc.counter, 1)
	index := int(counter) % len(bc.addresses)
	addr := bc.addresses[index]
	return bc.clients[addr]
}

// call will first get a rpc client by `client` function and use that client to request an RPC
func (bc *BitmarkdRPCClient) call(command string, args interface{}, reply interface{}) error {
	client := bc.client()
	err := client.Call(command, args, reply)
	if err != nil {
		if _, ok := err.(*net.OpError); ok {
			client.Close()
		}

		if err == rpc.ErrShutdown {
			client.Close()
		}
	}
	return err
}
