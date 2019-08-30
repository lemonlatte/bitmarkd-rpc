package bitmarkdClient

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/bitmark-inc/bitmark-sdk-go"
)

func setup() error {
	bitmarksdk.Init(&bitmarksdk.Config{
		APIToken: "bmk-lljpzkhqdkzmblhg",
		Network:  bitmarksdk.Testnet,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	})

	if err := os.Mkdir("test-certs", 0755); err != nil {
		return err
	}

	cmd := exec.Command("openssl",
		"req", "-new", "-nodes", "-x509", "-out", "test-certs/server.pem", "-keyout", "test-certs/server.key",
		"-days", "3650", "-subj", "/C=TW/ST=TPE/L=Earth/O=Bitmark Inc/OU=IT/CN=www.bitmark.com/emailAddress=test@bitmark.com")
	return cmd.Run()
}

func tearDown() {
	os.RemoveAll("test-certs")
}

func TestMain(m *testing.M) {
	err := setup()
	if err != nil {
		fmt.Println(err)
	}
	r := m.Run()
	tearDown()
	os.Exit(r)
}
