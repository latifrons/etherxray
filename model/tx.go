package model

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

type Tx struct {
	BasicTx    *types.Transaction
	MaxGasCost *big.Int
	From       common.Address
}
