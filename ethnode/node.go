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
		var receipt *types.Receipt
		receipt, err = n.RpcWrapper.BlockTxReceipts(tools.GetContextDefault(), tx.Hash())
		if err != nil {
			return
		}

		gasCost := big.NewInt(0).SetUint64(receipt.GasUsed)
		gasCost.Mul(gasCost, tx.GasPrice())
		sender, _ := types.Sender(n.Signer, tx)

		txs = append(txs, model.Tx{
			BasicTx: tx,
			Receipt: receipt,
			GasCost: gasCost,
			From:    sender,
			Rating:  receipt.GasUsed / 21000 / 5,
		})
	}
	return
}
