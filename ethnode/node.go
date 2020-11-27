package ethnode

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/latifrons/etherxray/middleware"
	"github.com/latifrons/etherxray/model"
	"github.com/latifrons/etherxray/tools"
	"math/big"
)

type EthNode struct {
	RpcWrapper *middleware.RpcWrapper
	Signer     types.EIP155Signer
}

func (n *EthNode) GetBlockTxs(height uint64) (txs []model.Tx, err error) {
	block, erro := n.RpcWrapper.BlockTxs(tools.GetContextDefault(), height)
	if erro != nil {
		err = erro
		return
	}
	for _, tx := range block.Transactions() {
		gasCost := big.NewInt(0).SetInt64(int64(tx.Gas()))
		gasCost.Mul(gasCost, tx.GasPrice())
		sender, _ := types.Sender(n.Signer, tx)

		txs = append(txs, model.Tx{
			BasicTx:    tx,
			MaxGasCost: gasCost,
			From:       sender,
		})
	}
	return
}
