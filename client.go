// Copyright (c) 2014-2017 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package bitmarkdClient

import (
	"crypto/tls"
	"fmt"
	"net/rpc"
	"net/rpc/jsonrpc"
	"sync"
	"time"
)

// ErrRPCConnection is an error for rpc connection
var ErrRPCConnection = fmt.Errorf("can not connect to rpc server")

// RPCEmptyArguments is an empty argument for rpc requests
type RPCEmptyArguments struct{}

// BitmarkdRPCClient is a struct for bitmarkd
type BitmarkdRPCClient struct {
	sync.Mutex
	client     *rpc.Client
	address    string
	connected  bool
	retryTimes uint
}

// NewBitmarkdRPCClient is to create a rpc client for bitmarkd
func New(address string) *BitmarkdRPCClient {
	client := &BitmarkdRPCClient{
		address:    address,
		retryTimes: 3,
	}
	return client
}

// Connect will establish rpc connection over tls
func (bc *BitmarkdRPCClient) Connect() error {
	bc.Lock()
	defer bc.Unlock()
	conn, err := tls.Dial("tcp", bc.address, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return err
	}
	bc.client = jsonrpc.NewClient(conn)
	bc.connected = true
	return nil
}

// Close will terminate the rpc connection
func (bc *BitmarkdRPCClient) Close() {
	bc.client.Close()
	bc.connected = false
}

func (bc *BitmarkdRPCClient) call(command string, args interface{}, reply interface{}) error {
	if bc.client == nil {
		err := bc.Connect()
		if err != nil {
			return ErrRPCConnection
		}
	}

	var err error
	for i := bc.retryTimes; i > 0; i-- {
		err = bc.client.Call(command, args, reply)

		if err != nil {
			if err == rpc.ErrShutdown {
				time.Sleep(time.Second)
				bc.connected = false
				bc.Connect()
				continue
			}
		}
		return err
	}
	return err
}

// Address returns the ip address of the rpc server
func (bc *BitmarkdRPCClient) Address() string {
	return bc.address
}
