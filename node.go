// Copyright (c) 2014-2017 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package bitmarkdClient

import (
	"encoding/json"
)

// GetNodeInfo will fetch node info from bitmarkd rpc server
func (bc *BitmarkdRPCClient) GetNodeInfo() (json.RawMessage, error) {
	args := RPCEmptyArguments{}
	var reply json.RawMessage
	err := bc.call("Node.Info", &args, &reply)
	return reply, err
}
