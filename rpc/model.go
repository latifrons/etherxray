package rpc

type RpcTx struct {
	Id         int    `json:"id"`
	Hash       string `json:"hash"`
	GasPrice   string `json:"gas_price"`
	MaxGasCost string `json:"max_gas_cost"`
	GasLimit   int    `json:"gas_limit"`
	GasUsed    int    `json:"gas_used"`
	From       string `json:"from"`
	To         string `json:"to"`
	Value      string `json:"value"`
	DataLength int    `json:"data_length"`
}
