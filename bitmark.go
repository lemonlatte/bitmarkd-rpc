// Copyright (c) 2014-2017 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package bitmarkdClient

import (
	"encoding/json"

	"github.com/bitmark-inc/bitmarkd/transactionrecord"
)

type CreateArguments struct {
	Assets []*transactionrecord.AssetData    `json:"assets"`
	Issues []*transactionrecord.BitmarkIssue `json:"issues"`
}

type ProofArguments struct {
	PayId string `json:"payId"`
	Nonce string `json:"nonce"`
}

// CreateBitmark includes two part: asset creating and issue creating
func (bc *BitmarkdRPCClient) CreateBitmark(assets []*transactionrecord.AssetData, issues []*transactionrecord.BitmarkIssue) (json.RawMessage, error) {
	var reply json.RawMessage

	args := CreateArguments{
		Assets: assets,
		Issues: issues,
	}

	err := bc.call("Bitmarks.Create", &args, &reply)
	return reply, err
}

// CreateIssue performs an issue creating request
func (bc *BitmarkdRPCClient) MakeProof(payId string, nonce string) (json.RawMessage, error) {
	var reply json.RawMessage

	args := ProofArguments{
		PayId: payId,
		Nonce: nonce,
	}

	err := bc.call("Bitmarks.Proof", &args, &reply)
	return reply, err
}
