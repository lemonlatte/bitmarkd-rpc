package bitmarkdClient

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"net/rpc"
	"net/rpc/jsonrpc"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestReply struct {
	OK int
}

type Dummy struct{}

func (d *Dummy) Test(args *RPCEmptyArguments, reply *TestReply) error {
	reply.OK = 1
	return nil
}

func (d *Dummy) Sleep(args *RPCEmptyArguments, reply *TestReply) error {
	time.Sleep(5 * time.Second)
	reply.OK = 1
	return nil
}

type Node struct {
	IP string
}
type TestNodeReply struct {
	IP string
}

func (n *Node) Info(args *RPCEmptyArguments, reply *TestNodeReply) error {
	reply.IP = n.IP
	return nil
}

func mockRPCServer(port string, handler interface{}) error {
	server := rpc.NewServer()
	err := server.Register(handler)
	if err != nil {
		return fmt.Errorf("Format of service Dummy isn't correct. %s", err)
	}

	cert, err := tls.LoadX509KeyPair("test-certs/server.pem", "test-certs/server.key")
	if err != nil {
		return fmt.Errorf("server: loadkeys: %s", err)
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	config.Rand = rand.Reader

	l, err := tls.Listen("tcp", ":"+port, &config)
	if err != nil {
		return fmt.Errorf("Couldn't start listening on port 1234. Error %s", err)
	}

	go func() {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		server.ServeCodec(jsonrpc.NewServerCodec(conn))
	}()
	return nil
}

func TestPersistentRPCClientRequestSuccessfully(t *testing.T) {
	assert.NoError(t, mockRPCServer("12345", new(Dummy)))

	c := NewPersistentRPCClient("127.0.0.1:12345", 10*time.Second)
	args := RPCEmptyArguments{}
	var reply TestReply

	err := c.Call("Dummy.Test", args, &reply)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, reply.OK)
}

func TestPersistentRPCClientConnectFailure(t *testing.T) {
	c := NewPersistentRPCClient("127.0.0.1:12346", time.Second)
	args := RPCEmptyArguments{}
	var reply TestReply

	err := c.Call("Dummy.Test", args, &reply)
	assert.EqualError(t, err, `connection can not be established`)
}

func TestPersistentRPCClientConnectTimeout(t *testing.T) {
	assert.NoError(t, mockRPCServer("12347", new(Dummy)))

	c := NewPersistentRPCClient("127.0.0.1:12347", time.Millisecond)
	args := RPCEmptyArguments{}
	var reply TestReply

	err := c.Call("Dummy.Sleep", args, &reply)
	assert.EqualError(t, err, ErrRPCRequestTimeout.Error())
}

func TestPersistentRPCClientConnectClosed(t *testing.T) {
	assert.NoError(t, mockRPCServer("12348", new(Dummy)))

	c := NewPersistentRPCClient("127.0.0.1:12348", 5*time.Second)
	args := RPCEmptyArguments{}
	var reply TestReply

	go func() {
		time.Sleep(10 * time.Millisecond)
		c.Close()
	}()

	err := c.Call("Dummy.Sleep", args, &reply)
	assert.EqualError(t, err, ErrRPCConnectionClosed.Error())
}

func TestBitmarkdRPCClientDummyNodeInfo(t *testing.T) {
	assert.NoError(t, mockRPCServer("23456", &Node{IP: "localhost:23456"}))
	assert.NoError(t, mockRPCServer("23457", &Node{IP: "localhost:23457"}))

	c := New([]string{"127.0.0.1:23456", "127.0.0.1:23457"}, 5*time.Second)

	data, err := c.GetNodeInfo()
	assert.NoError(t, err)
	assert.Equal(t, `{"IP":"localhost:23457"}`, string(data))

	data, err = c.GetNodeInfo()
	assert.NoError(t, err)
	assert.Equal(t, `{"IP":"localhost:23456"}`, string(data))
}
