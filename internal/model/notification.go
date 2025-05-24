package model

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type NotiEvent struct {
	DappOwner common.Address `json:"dappOwner"`
	User      common.Address `json:"user"`
	EventId   *big.Int
	Title     string   `json:"title"`
	Message   string   `json:"message"`
	CreatedAt *big.Int `json:"createdAt"`
	// SystemApp bool           `json:"systemApp"`
}

// type Notification struct {
// 	Dapp        string `json:"dapp"`
// 	User        string `json:"user"`
// 	Title       string `json:"title"`
// 	Body        string `json:"body"`
// 	AtTime      uint64 `json:"atTime"`
// 	DeviceToken string `json:"deviceToken"`
// }
