package redis

import (
	"log"
	"os"
	"testing"

	"github.com/barpilot/gosba/crypto"
	"github.com/barpilot/gosba/crypto/noop"
)

func TestMain(m *testing.M) {
	if err := crypto.InitializeGlobalCodec(noop.NewCodec()); err != nil {
		log.Fatal(err)
	}
	os.Exit(m.Run())
}
