package constellation

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/patrickmn/go-cache"
	"time"
)

func copyBytes(b []byte) []byte {
	ob := make([]byte, len(b))
	copy(ob, b)
	return ob
}

type Constellation struct {
	node             *Client
	c                *cache.Cache
	maskAddress      common.Address
	nullAddressProxy common.Address
}

func (g *Constellation) Send(toField common.Address, data []byte, from string, to []string) (out []byte, err error) {
	payload := append(toField.Bytes(), data...)
	if len(data) > 0 {
		if len(to) == 0 {
			out = copyBytes(payload)
		} else {
			var err error
			out, err = g.node.SendPayload(payload, from, to)
			if err != nil {
				return nil, err
			}
		}
	}
	g.c.Set(string(out), payload, cache.DefaultExpiration)
	return out, nil
}

func (g *Constellation) ParseConstellationPayload(dataWithTo []byte) (*common.Address, []byte) {
	realTo := common.BytesToAddress(dataWithTo[:20])
	realPayload := dataWithTo[20:]
	if realTo != g.nullAddressProxy {
		return &realTo, realPayload
	} else {
		return nil, realPayload
	}

}

func (g *Constellation) Receive(data []byte) (*common.Address, []byte, error) {
	dataStr := string(data)
	x, found := g.c.Get(dataStr)
	if found {
		realTo, dat := g.ParseConstellationPayload(x.([]byte))
		return realTo, dat, nil
	}
	// Ignore this error since not being a recipient of
	// a payload isn't an error.
	// TODO: Return an error if it's anything OTHER than
	// 'you are not a recipient.'
	dataWithTo, _ := g.node.ReceivePayload(data)
	realTo, realData := g.ParseConstellationPayload(dataWithTo)
	g.c.Set(dataStr, dataWithTo, cache.DefaultExpiration)
	return realTo, realData, nil
}

func (g *Constellation) MaskTo(real **common.Address) {
	*real = &g.maskAddress
}

func New(configPath string) (*Constellation, error) {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}
	err = RunNode(configPath, cfg.Socket)
	if err != nil {
		return nil, err
	}
	n, err := NewClient(cfg.PublicKeys[0], cfg.Socket)
	if err != nil {
		return nil, err
	}
	maskAddr := common.BytesToAddress(common.FromHex(cfg.ToMask))
	nullProxy := common.BytesToAddress(common.FromHex(cfg.NullProxy))
	return &Constellation{
		node:             n,
		c:                cache.New(5*time.Minute, 5*time.Minute),
		maskAddress:      maskAddr,
		nullAddressProxy: nullProxy,
	}, nil
}

func MustNew(configPath string) *Constellation {
	g, err := New(configPath)
	if err != nil {
		panic(fmt.Sprintf("MustNew error: %v", err))
	}
	return g
}

func MaybeNew(configPath string) *Constellation {
	if configPath == "" {
		return nil
	}
	return MustNew(configPath)
}
