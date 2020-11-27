package core

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/latifrons/etherxray/ethnode"
	"github.com/latifrons/etherxray/middleware"
	"github.com/latifrons/etherxray/rpc"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"math/big"
)

type Node struct {
	DataFolder string

	components []Component
}

func (n *Node) Setup() {
	signer := types.NewEIP155Signer(big.NewInt(1))

	rpcWrapper := &middleware.RpcWrapper{
		RpcAddress: viper.GetString("node.address"),
	}

	ethNode := &ethnode.EthNode{
		RpcWrapper: rpcWrapper,
		Signer:     signer,
	}

	rpcServer := &rpc.RpcServer{
		C: &rpc.RpcController{
			EthNode: ethNode,
		},
		Port: viper.GetString("rpc.port"),
	}
	rpcServer.InitDefault()

	n.components = append(n.components, rpcServer)
}

func (n *Node) Start() {
	for _, component := range n.components {
		logrus.Infof("Starting %s", component.Name())
		component.Start()
		logrus.Infof("Started: %s", component.Name())

	}
	//n.heightEventChan <- 10943851
	logrus.Info("Node Started")
}

func (n *Node) Stop() {
	for i := len(n.components) - 1; i >= 0; i-- {
		comp := n.components[i]
		logrus.Infof("Stopping %s", comp.Name())
		comp.Stop()
		logrus.Infof("Stopped: %s", comp.Name())
	}
	logrus.Info("Node Stopped")
}
