package rpc

type RpcTx struct {
	Id         int    `json:"id"`
	Success    bool   `json:"success"`
	Hash       string `json:"hash"`
	GasPrice   string `json:"gas_price"`
	GasLimit   uint64 `json:"gas_limit"`
	GasUsed    uint64 `json:"gas_used"`
	GasCost    string `json:"gas_cost"`
	From       string `json:"from"`
	To         string `json:"to"`
	Value      string `json:"value"`
	DataLength int    `json:"data_length"`
	Rating     uint64 `json:"rating"`
}
