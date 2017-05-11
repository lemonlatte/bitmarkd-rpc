// Copyright (c) 2014-2017 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package bitmarkdClient

import (
	"encoding/json"
)

// TransactionArguments represents a transaction
type TransactionArguments struct {
	Id string `json:"txId"`
}

// GetTransactionStatus will return the status for a specific transaction id.
func (bc *BitmarkdRPCClient) GetTransactionStatus(txId string) (json.RawMessage, error) {
	args := TransactionArguments{Id: txId}
	var reply json.RawMessage
	err := bc.call("Transaction.Status", &args, &reply)
	return reply, err
}
