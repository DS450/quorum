package private

import (
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/private/constellation"
)

type PrivateTransactionManager interface {
	Send(realTo common.Address, data []byte, from string, to []string) ([]byte, error)
	Receive(data []byte) (*common.Address, []byte, error)
	NullAddressProxy() common.Address
	ParseConstellationPayload(data []byte) (realTo *common.Address, realData []byte)
	MaskTo(**common.Address)
}

func FromEnvironmentOrNil(name string) PrivateTransactionManager {
	cfgPath := os.Getenv(name)
	if cfgPath == "" {
		return nil
	}
	return constellation.MustNew(cfgPath)
}

var P = FromEnvironmentOrNil("PRIVATE_CONFIG")
