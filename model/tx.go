package model

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

type Tx struct {
	BasicTx *types.Transaction
	Receipt *types.Receipt
	From    common.Address
	GasCost *big.Int
	Rating  uint64
}
