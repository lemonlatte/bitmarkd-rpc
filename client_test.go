package bitmarkdClient

import (
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	"golang.org/x/crypto/sha3"

	sdkAccount "github.com/bitmark-inc/bitmark-sdk-go/account"
	sdkAsset "github.com/bitmark-inc/bitmark-sdk-go/asset"
	sdkBitmark "github.com/bitmark-inc/bitmark-sdk-go/bitmark"
	"github.com/bitmark-inc/bitmarkd/transactionrecord"
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

type RewriteIssue struct {
	AssetId   string `json:"assetId" pack:"hex64"`
	Owner     string `json:"owner" pack:"account"`
	Nonce     uint64 `json:"nonce" pack:"uint64"`
	Signature string `json:"signature"`
}

func benchmarkCreateIssue(b *testing.B, c *BitmarkdRPCClient) {
	creator, err := sdkAccount.FromSeed("5XEECsgiSHnF2AVaEsJ9hkSrjrcozdyecp7BReVyesi3mVnfahsqyQZ")
	if err != nil {
		panic(err)
	}

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		a, _ := sdkAsset.NewRegistrationParams("test", nil)
		a.SetFingerprint([]byte(time.Now().String()))
		a.Sign(creator)

		digest := sha3.Sum512([]byte(a.Fingerprint))
		assetID := hex.EncodeToString(digest[:])

		i := sdkBitmark.NewIssuanceParams(assetID, 1)
		i.Sign(creator)

		assetByte, _ := json.Marshal(a)
		issueByte, _ := json.Marshal(RewriteIssue(*i.Issuances[0]))

		var newAsset *transactionrecord.AssetData
		var newIssue *transactionrecord.BitmarkIssue
		if err := json.Unmarshal(assetByte, &newAsset); err != nil {
			b.Fatal(string(assetByte), err)
		}
		if err := json.Unmarshal(issueByte, &newIssue); err != nil {
			b.Fatal(string(issueByte), err)
		}

		b.StartTimer()
		_, err = c.CreateBitmark([]*transactionrecord.AssetData{newAsset}, []*transactionrecord.BitmarkIssue{newIssue})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCreateIssue(b *testing.B) {
	benchmarks := map[string]string{
		"LOCAL": "127.0.0.1:2230",
		"DO":    "node-d1.test.bitmark.com:2130",
		"AWS":   "54.95.248.16:2130",
	}

	for name, ip := range benchmarks {
		b.Run(name, func(b *testing.B) {
			c := New(ip)
			benchmarkCreateIssue(b, c)
		})
	}
}
