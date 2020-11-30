package rpc

import (
	"errors"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/latifrons/etherxray/ethnode"
	"github.com/latifrons/etherxray/model"
	"github.com/latifrons/etherxray/tools"
	"github.com/sirupsen/logrus"
	"math/big"
	"net/http"
	"time"
)

type RpcController struct {
	EthNode *ethnode.EthNode
}

func (rpc *RpcController) NewRouter() *gin.Engine {
	router := gin.New()
	if logrus.GetLevel() > logrus.DebugLevel {
		logger := gin.LoggerWithConfig(gin.LoggerConfig{
			Formatter: ginLogFormatter,
			Output:    logrus.StandardLogger().Out,
			SkipPaths: []string{"/"},
		})
		router.Use(logger)
	}

	router.Use(gin.RecoveryWithWriter(logrus.StandardLogger().Out))
	return rpc.addRouter(router)
}

var ginLogFormatter = func(param gin.LogFormatterParams) string {
	if logrus.GetLevel() < logrus.TraceLevel {
		return ""
	}
	var statusColor, methodColor, resetColor string
	if param.IsOutputColor() {
		statusColor = param.StatusCodeColor()
		methodColor = param.MethodColor()
		resetColor = param.ResetColor()
	}

	if param.Latency > time.Minute {
		// Truncate in a golang < 1.8 safe way
		param.Latency = param.Latency - param.Latency%time.Second
	}

	logEntry := fmt.Sprintf("GIN %v %s %3d %s %13v  %15s %s %-7s %s %s %s",
		param.TimeStamp.Format("2006/01/02 - 15:04:05"),
		statusColor, param.StatusCode, resetColor,
		param.Latency,
		param.ClientIP,
		methodColor, param.Method, resetColor,
		param.Path,
		param.ErrorMessage,
	)
	logrus.Tracef("gin log %v ", logEntry)
	//return  logEntry
	return ""
}

func (rpc *RpcController) addRouter(router *gin.Engine) *gin.Engine {
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	router.Use(static.Serve("/", static.LocalFile("web", false)))

	router.GET("/health", rpc.Health)
	router.GET("/block/:height", rpc.Block)

	return router
}

func Response(c *gin.Context, status int, err error, data interface{}) {
	c.JSON(status, data)
}

func (rpc *RpcController) Health(c *gin.Context) {
	Response(c, http.StatusOK, nil, "ok")
}

func (rpc *RpcController) Block(c *gin.Context) {
	heightS := c.Param("height")

	height, ok := big.NewInt(0).SetString(heightS, 10)
	if !ok {
		Response(c, http.StatusBadRequest, errors.New("bad height"), nil)
		return
	}
	txs, err := rpc.EthNode.GetBlockTxs(height.Uint64())
	if err != nil {
		Response(c, http.StatusInternalServerError, err, nil)
		return
	}
	txsm := rpc.toRpcTxs(txs)

	Response(c, http.StatusOK, nil, txsm)
	return
}

func (rpc *RpcController) toRpcTxs(txs []model.Tx) (rpcTx []RpcTx) {
	for i, tx := range txs {
		var to string
		if tx.BasicTx.To() != nil {
			to = tx.BasicTx.To().Hex()
		} else {
			to = ""
		}

		rpcTx = append(rpcTx, RpcTx{
			Id:         i,
			Success:    tx.Receipt.Status == 1,
			Hash:       tx.BasicTx.Hash().Hex(),
			GasPrice:   tools.FromWeiToGwei(tx.BasicTx.GasPrice()).FloatString(8),
			GasCost:    tools.FromWei(tx.GasCost).FloatString(8),
			GasLimit:   tx.BasicTx.Gas(),
			GasUsed:    tx.Receipt.GasUsed,
			From:       tx.From.Hex(),
			To:         to,
			Value:      tools.FromWei(tx.BasicTx.Value()).FloatString(8),
			DataLength: len(tx.BasicTx.Data()),
			Rating:     tx.Rating,
		})
	}
	return

}
