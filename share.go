// Copyright (c) 2014-2017 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package bitmarkdClient

import (
	"encoding/json"

	"github.com/bitmark-inc/bitmarkd/transactionrecord"
)

// CreateBitmark includes two part: asset creating and issue creating
func (bc *BitmarkdRPCClient) CreateShares(share transactionrecord.BitmarkShare) (json.RawMessage, error) {
	var reply json.RawMessage
	err := bc.call("Share.Create", &share, &reply)
	return reply, err
}

func (bc *BitmarkdRPCClient) GrantShares(grant transactionrecord.ShareGrant) (json.RawMessage, error) {
	var reply json.RawMessage
	err := bc.call("Share.Grant", &grant, &reply)
	return reply, err
}

func (bc *BitmarkdRPCClient) SwapShares(swap transactionrecord.ShareSwap) (json.RawMessage, error) {
	var reply json.RawMessage
	err := bc.call("Share.Swap", &swap, &reply)
	return reply, err
}
