package middleware

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fastjson"
	"math/big"
	"strings"
)

var Zero = big.NewInt(0)

var (
	TyUint256, _ = abi.NewType("uint256", "", nil)
	TyUint32, _  = abi.NewType("uint32", "", nil)
	TyUint16, _  = abi.NewType("uint16", "", nil)
	//TyUint, _       = abi.NewType("uint", "", nil)
	TyString, _     = abi.NewType("string", "", nil)
	TyBool, _       = abi.NewType("bool", "", nil)
	TyByte, _       = abi.NewType("byte", "", nil)
	TyBytes, _      = abi.NewType("bytes", "", nil)
	TyByteArr, _    = abi.NewType("bytes1[]", "", nil)
	TyBytes32, _    = abi.NewType("bytes32", "", nil)
	TyAddress, _    = abi.NewType("address", "", nil)
	TyUint64Arr, _  = abi.NewType("uint64[]", "", nil)
	TyAddressArr, _ = abi.NewType("address[]", "", nil)
	TyInt8, _       = abi.NewType("int8", "", nil)
	// Special types for testing
	TyUint32Arr2, _       = abi.NewType("uint32[2]", "", nil)
	TyUint64Arr2, _       = abi.NewType("uint64[2]", "", nil)
	TyUint256Arr, _       = abi.NewType("uint256[]", "", nil)
	TyUint256Arr2, _      = abi.NewType("uint256[2]", "", nil)
	TyUint256Arr3, _      = abi.NewType("uint256[3]", "", nil)
	TyUint256ArrNested, _ = abi.NewType("uint256[2][2]", "", nil)
	TyUint8ArrNested, _   = abi.NewType("uint8[][2]", "", nil)
	TyUint8SliceNested, _ = abi.NewType("uint8[][]", "", nil)
	TyTupleF, _           = abi.NewType("tuple", "struct Overloader.F", []abi.ArgumentMarshaling{
		{Name: "_f", Type: "uint256"},
		{Name: "__f", Type: "uint256"},
		{Name: "f", Type: "uint256"}})
)

var methods = map[string]abi.Method{
	"_general_list_address": abi.NewMethod("_general_list_address", "_general_list_address", abi.Function, "", false, false,
		[]abi.Argument{
			{"index", TyUint256, false},
		},
		[]abi.Argument{{"address", TyAddress, false}}),
	"_general_uint256": abi.NewMethod("_general_uint256", "_general_uint256", abi.Function, "", false, false,
		[]abi.Argument{},
		[]abi.Argument{{"value", TyUint256, false}}),
	"_general_bytes32": abi.NewMethod("_general_bytes32", "_general_bytes32", abi.Function, "", false, false,
		[]abi.Argument{},
		[]abi.Argument{{"value", TyBytes32, false}}),
	"_general_string": abi.NewMethod("_general_string", "_general_string", abi.Function, "", false, false,
		[]abi.Argument{},
		[]abi.Argument{{"value", TyString, false}}),
	"_general_bool": abi.NewMethod("_general_bool", "_general_bool", abi.Function, "", false, false,
		[]abi.Argument{},
		[]abi.Argument{{"value", TyBool, false}}),
	"_general_address": abi.NewMethod("_general_address", "_general_address", abi.Function, "", false, false,
		[]abi.Argument{},
		[]abi.Argument{{"value", TyAddress, false}}),
	"_general_address_array": abi.NewMethod("_general_address_array", "_general_address_array", abi.Function, "", false, false,
		[]abi.Argument{},
		[]abi.Argument{{"value", TyAddressArr, false}}),
	"getPair": abi.NewMethod("getPair", "getPair", abi.Function, "", false, false,
		[]abi.Argument{
			{"token1", TyAddress, false},
			{"token2", TyAddress, false},
		},
		[]abi.Argument{{"pairAddress", TyAddress, false}}),

	"allPairsLength": abi.NewMethod("allPairsLength", "allPairsLength", abi.Function, "", false, false,
		[]abi.Argument{},
		[]abi.Argument{{"length", TyUint256, false}}),
	"getReserves": abi.NewMethod("getReserves", "getReserves", abi.Function, "", false, false,
		[]abi.Argument{},
		[]abi.Argument{
			{"reserve0", TyUint256, false},
			{"reserve1", TyUint256, false},
			{"timestamp", TyUint256, false},
		}),
	"getDenormalizedWeight": abi.NewMethod("getDenormalizedWeight", "getDenormalizedWeight", abi.Function, "", false, false,
		// input
		[]abi.Argument{{"token", TyAddress, false}},
		// output
		[]abi.Argument{{"weight", TyUint256, false}},
	),
	"getBalance": abi.NewMethod("getBalance", "getBalance", abi.Function, "", false, false,
		// input
		[]abi.Argument{{"token", TyAddress, false}},
		// output
		[]abi.Argument{{"balance", TyUint256, false}},
	),
	"getSwapFee": abi.NewMethod("getSwapFee", "getSwapFee", abi.Function, "", false, false,
		// input
		[]abi.Argument{{"token", TyAddress, false}},
		// output
		[]abi.Argument{{"swapFee", TyUint256, false}},
	),
}

type GetReservesResponse struct {
	Reserve0  *big.Int
	Reserve1  *big.Int
	Timestamp *big.Int
}

var myAbi = abi.ABI{
	Methods: methods,
}

type RpcWrapper struct {
	RpcAddress         string
	Signer             types.Signer
	MaxTxAllowedToSend int

	sent int
}

func (r *RpcWrapper) BlockHeight(ctx context.Context) (uint64, error) {
	client, _ := ethclient.DialContext(ctx, r.RpcAddress)
	defer client.Close()

	return client.BlockNumber(ctx)
}

func (r *RpcWrapper) BlockTxs(ctx context.Context, height uint64) (*types.Block, error) {
	client, _ := ethclient.DialContext(ctx, r.RpcAddress)
	defer client.Close()

	return client.BlockByNumber(ctx, big.NewInt(0).SetUint64(height))
}

func (r *RpcWrapper) GetValueRetUint(ctx context.Context, contract common.Address, field string) (int2 *big.Int, err error) {
	method := crypto.Keccak256([]byte(field + "()"))[:4]

	bytes, err := myAbi.Pack("_general_uint256")
	if err != nil {
		logrus.WithError(err).Error("pack field")
		return
	}
	allBytes := append(method, bytes[4:]...)

	client, _ := ethclient.DialContext(ctx, r.RpcAddress)

	ret, err := client.CallContract(ctx, ethereum.CallMsg{
		From:     common.Address{},
		To:       &contract,
		Gas:      0,
		GasPrice: Zero,
		Value:    Zero,
		Data:     allBytes,
	}, nil)
	if err != nil {
		logrus.WithError(err).Warn("call contract")
		return
	}
	err = myAbi.UnpackIntoInterface(&int2, "_general_uint256", ret)
	return
}

func (r *RpcWrapper) GetValueRetString(ctx context.Context, contract common.Address, field string) (value string, err error) {
	method := crypto.Keccak256([]byte(field + "()"))[:4]

	bytesRaw, err := myAbi.Pack("_general_string")
	if err != nil {
		logrus.WithError(err).Error("pack field")
		return
	}
	allBytes := append(method, bytesRaw[4:]...)

	client, _ := ethclient.DialContext(ctx, r.RpcAddress)

	ret, err := client.CallContract(ctx, ethereum.CallMsg{
		From:     common.Address{},
		To:       &contract,
		Gas:      0,
		GasPrice: Zero,
		Value:    Zero,
		Data:     allBytes,
	}, nil)
	if err != nil {
		logrus.WithError(err).Warn("call contract")
		return
	}
	err = myAbi.UnpackIntoInterface(&value, "_general_string", ret)
	if err != nil {
		bts := [32]byte{}
		err = myAbi.UnpackIntoInterface(&bts, "_general_bytes32", ret)
		if err == nil {
			bytesStr := bytes.Trim(bts[:], "\x00")
			value = string(bytesStr)
		}
	}
	value = strings.TrimSpace(value)
	return
}

func (r *RpcWrapper) GetValueRetAddress(ctx context.Context, contract common.Address, field string) (addr common.Address, err error) {
	method := crypto.Keccak256([]byte(field + "()"))[:4]

	bytes, err := myAbi.Pack("_general_address")
	if err != nil {
		logrus.WithError(err).Error("pack field")
		return
	}
	allBytes := append(method, bytes[4:]...)

	client, _ := ethclient.DialContext(ctx, r.RpcAddress)

	ret, err := client.CallContract(ctx, ethereum.CallMsg{
		From:     common.Address{},
		To:       &contract,
		Gas:      0,
		GasPrice: Zero,
		Value:    Zero,
		Data:     allBytes,
	}, nil)
	if err != nil {
		logrus.WithError(err).Warn("call contract")
		return
	}
	err = myAbi.UnpackIntoInterface(&addr, "_general_address", ret)
	return
}

func (r *RpcWrapper) GetValueRetBool(ctx context.Context, contract common.Address, field string) (value bool, err error) {
	method := crypto.Keccak256([]byte(field + "()"))[:4]

	bytes, err := myAbi.Pack("_general_bool")
	if err != nil {
		logrus.WithError(err).Error("pack field")
		return
	}
	allBytes := append(method, bytes[4:]...)

	client, _ := ethclient.DialContext(ctx, r.RpcAddress)

	ret, err := client.CallContract(ctx, ethereum.CallMsg{
		From:     common.Address{},
		To:       &contract,
		Gas:      0,
		GasPrice: Zero,
		Value:    Zero,
		Data:     allBytes,
	}, nil)
	if err != nil {
		logrus.WithError(err).Warn("call contract")
		return
	}
	err = myAbi.UnpackIntoInterface(&value, "_general_bool", ret)
	return
}

func (r *RpcWrapper) GetValueRetAddressArray(ctx context.Context, contract common.Address, field string) (addr []common.Address, err error) {
	method := crypto.Keccak256([]byte(field + "()"))[:4]

	bytes, err := myAbi.Pack("_general_address_array")
	if err != nil {
		logrus.WithError(err).Error("pack field")
		return
	}
	allBytes := append(method, bytes[4:]...)

	client, _ := ethclient.DialContext(ctx, r.RpcAddress)

	ret, err := client.CallContract(ctx, ethereum.CallMsg{
		From:     common.Address{},
		To:       &contract,
		Gas:      0,
		GasPrice: Zero,
		Value:    Zero,
		Data:     allBytes,
	}, nil)
	if err != nil {
		logrus.WithError(err).Warn("call contract")
		return
	}
	err = myAbi.UnpackIntoInterface(&addr, "_general_address_array", ret)
	return
}

func (r *RpcWrapper) GetListValueByIndexRetAddress(ctx context.Context, contract common.Address, mapName string, index int) (addr common.Address, err error) {
	method := crypto.Keccak256([]byte(mapName + "(uint256)"))[:4]

	bytes, err := myAbi.Pack("_general_list_address", big.NewInt(int64(index)))
	if err != nil {
		logrus.WithError(err).Error("pack field")
		return
	}
	allBytes := append(method, bytes[4:]...)

	client, _ := ethclient.DialContext(ctx, r.RpcAddress)

	ret, err := client.CallContract(ctx, ethereum.CallMsg{
		From:     common.Address{},
		To:       &contract,
		Gas:      0,
		GasPrice: Zero,
		Value:    Zero,
		Data:     allBytes,
	}, nil)
	if err != nil {
		logrus.WithError(err).Warn("call contract")
		return
	}
	err = myAbi.UnpackIntoInterface(&addr, "_general_list_address", ret)
	return
}

func (r *RpcWrapper) GetUniswapLiquidities(ctx context.Context, contract common.Address, height *big.Int) (resp *GetReservesResponse, err error) {
	bytes, err := myAbi.Pack("getReserves")
	if err != nil {
		logrus.WithError(err).Error("pack field")
		return
	}

	client, _ := ethclient.DialContext(ctx, r.RpcAddress)

	ret, err := client.CallContract(ctx, ethereum.CallMsg{
		From:     common.Address{},
		To:       &contract,
		Gas:      0,
		GasPrice: Zero,
		Value:    Zero,
		Data:     bytes,
	}, height)
	if err != nil {
		logrus.WithError(err).Warn("call contract")
		return
	}

	resp = &GetReservesResponse{}

	err = myAbi.UnpackIntoInterface(resp, "getReserves", ret)
	return
}

// Sync log
func (r *RpcWrapper) GetTradeLog(ctx context.Context, height uint64, topics []common.Hash) (logs []types.Log, err error) {
	client, _ := ethclient.DialContext(ctx, r.RpcAddress)
	logs, err = client.FilterLogs(ctx, ethereum.FilterQuery{
		BlockHash: nil,
		FromBlock: big.NewInt(int64(height)),
		ToBlock:   big.NewInt(int64(height)),
		//Addresses: nil,
		Topics: [][]common.Hash{topics},
	})
	return
}

func (r *RpcWrapper) GetTradeLogFromTo(ctx context.Context, fromHeight uint64, toHeight uint64, topic common.Hash, addresses []common.Address) (logs []types.Log, err error) {
	client, _ := ethclient.DialContext(ctx, r.RpcAddress)
	logs, err = client.FilterLogs(ctx, ethereum.FilterQuery{
		BlockHash: nil,
		FromBlock: big.NewInt(int64(fromHeight)),
		ToBlock:   big.NewInt(int64(toHeight)),
		Addresses: addresses,
		Topics:    [][]common.Hash{{topic}},
	})
	return
}

func (r *RpcWrapper) GetTradeLogs(ctx context.Context, height uint64, topics []common.Hash) (logs []types.Log, err error) {
	client, _ := ethclient.DialContext(ctx, r.RpcAddress)
	logs, err = client.FilterLogs(ctx, ethereum.FilterQuery{
		BlockHash: nil,
		FromBlock: big.NewInt(int64(height)),
		ToBlock:   big.NewInt(int64(height)),
		//Addresses: nil,
		Topics: [][]common.Hash{topics},
	})
	return
}

func (r *RpcWrapper) GetBlockGasPrices(ctx context.Context, height uint64) (gases []uint64, err error) {
	client, _ := ethclient.DialContext(ctx, r.RpcAddress)
	block, er := client.BlockByNumber(ctx, big.NewInt(int64(height)))
	if er != nil {
		err = er
		return
	}
	for _, tx := range block.Transactions() {
		gases = append(gases, tx.GasPrice().Uint64())
		//logrus.WithField("height", height).Info(tx.GasPrice().String())
	}
	return
}

func (r *RpcWrapper) GetSuggestedGasPrice(ctx context.Context) (*big.Int, error) {
	client, _ := ethclient.DialContext(ctx, r.RpcAddress)
	return client.SuggestGasPrice(ctx)
}

func (r *RpcWrapper) GetTxPoolRaw(ctx context.Context) (obj *fastjson.Object, err error) {
	c, err := rpc.DialContext(ctx, r.RpcAddress)
	if err != nil {
		return nil, err
	}
	var response json.RawMessage

	err = c.CallContext(ctx, &response, "txpool_content")
	if err != nil {
		return nil, err
	}
	bytes, _ := response.MarshalJSON()

	v, err := fastjson.ParseBytes(bytes)
	if err != nil {
		return nil, err
	}

	obj = v.GetObject("pending")

	return obj, nil
}

func (r *RpcWrapper) PendingNonceAt(ctx context.Context, address common.Address) (nonce uint64, err error) {
	client, _ := ethclient.DialContext(ctx, r.RpcAddress)

	nonce, err = client.PendingNonceAt(ctx, address)
	if err != nil {
		logrus.WithError(err).Error("failed to get nonce")
	}
	return
}
func (r *RpcWrapper) NonceAt(ctx context.Context, address common.Address) (nonce uint64, err error) {
	client, _ := ethclient.DialContext(ctx, r.RpcAddress)

	nonce, err = client.NonceAt(ctx, address, nil)
	if err != nil {
		logrus.WithError(err).Error("failed to get nonce")
	}
	return
}

func HexToAccount(hexPrivKey string) (*ecdsa.PrivateKey, *ecdsa.PublicKey, common.Address, error) {
	priK, err := crypto.HexToECDSA(hexPrivKey)
	if err != nil {
		return nil, nil, common.Address{}, err
	}
	pubK := priK.Public()
	pubKEcdsa := pubK.(*ecdsa.PublicKey)
	address := crypto.PubkeyToAddress(priK.PublicKey)

	return priK, pubKEcdsa, address, nil
}

func (r *RpcWrapper) GetBalanceETH(ctx context.Context, account common.Address) (v *big.Int, err error) {
	client, _ := ethclient.DialContext(ctx, r.RpcAddress)
	return client.BalanceAt(ctx, account, nil)
}

func (r *RpcWrapper) GetTransactionByHash(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error) {
	client, _ := ethclient.DialContext(ctx, r.RpcAddress)
	return client.TransactionByHash(ctx, hash)
}

func (r *RpcWrapper) GetTx(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error) {
	client, _ := ethclient.DialContext(ctx, r.RpcAddress)
	defer client.Close()
	return client.TransactionByHash(ctx, hash)
}
