package tools

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"math/rand"
	"time"
)

var e18 = big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(18)), nil)
var e9 = big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(9)), nil)
var e18r = big.NewRat(1, 1).SetInt(e18)

func StrToNum(height string) (value uint64, err error) {
	n := new(big.Int)
	n, ok := n.SetString(height, 0)
	if !ok {
		err = errors.New("bad height format")
		return
	}
	return n.Uint64(), nil
}

func MustStrToNum(v string) (value uint64) {
	n := new(big.Int)
	n, ok := n.SetString(v, 0)
	if !ok {
		panic("bad v format:" + v)
	}

	return n.Uint64()
}

func MustStrToBigInt(v string) *big.Int {
	n := new(big.Int)
	n, ok := n.SetString(v, 0)
	if !ok {
		panic("bad v format:" + v)
	}
	return n
}

func NumToStr(height uint64) string {
	return fmt.Sprintf("0x%X", height)
}

func ToWei(value string) *big.Int {
	inputValue, ok := new(big.Rat).SetString(value)
	if !ok {
		panic("failed to convert")
	}
	inputValue.Mul(inputValue, e18r)

	intv := big.NewInt(0).Set(inputValue.Num())
	intv.Quo(intv, inputValue.Denom())

	return intv
}

func GasToWei(gasPrice uint64) *big.Int {
	v := big.NewInt(0).SetUint64(gasPrice)
	v.Mul(v, e9)
	return v
}

func FromWei(value *big.Int) *big.Rat {
	return big.NewRat(1, 1).SetFrac(value, e18)
}

func FromWeiToGwei(value *big.Int) *big.Rat {
	return big.NewRat(1, 1).SetFrac(value, e9)
}

func RandomBetween(left *big.Int, right *big.Int) *big.Int {
	rate := rand.Float64()
	diff := big.NewInt(0).Set(right)
	diff.Sub(diff, left)
	c := big.NewRat(1, 1).SetInt(diff)
	d := big.NewRat(1, 1).SetFloat64(rate)
	c.Mul(c, d)
	result := big.NewInt(0).Set(c.Num())
	result.Quo(result, c.Denom())
	return result
}

var Barrier = 32

func CutInput(data []byte, hasMethod bool) (slots [][]byte, err error) {
	if len(data) < 4 {
		err = errors.New("bad data")
		return
	}
	pos := 0
	if hasMethod {
		slots = append(slots, data[0:4])
		pos = 4
	}
	for pos < len(data) {
		slots = append(slots, data[pos:pos+Barrier])
		pos += Barrier
	}
	return slots, nil
}

func GetContext(second int) context.Context {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*time.Duration(second))
	return ctx
}

func GetContextDefault() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	return ctx
}

func GasPriceWeiToString(gas *big.Int) string {
	str := gas.String()
	l := len(str)
	if l <= 9 {
		return str
	}
	return str[0:l-9] + "," + str[l-9:]

}

func SafeToString(to *common.Address) string {
	if to == nil {
		return "NULL"
	}
	return to.String()
}
