/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/ledger"
	"github.com/hyperledger/fabric/core/peer"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// New returns an implementation of the chaincode interface
func New() *GeneralQuerier {
	return &GeneralQuerier{}
}

// GeneralQuerier implements cross-chaincode ledger query
type GeneralQuerier struct{}

var logger = flogging.MustGetLogger("gscc")

// These are function names from Invoke first parameter
const (
	GetState string = "GetState"
)

// Init implements the chaincode shim interface
func (s *GeneralQuerier) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("Init GSCC")

	return shim.Success(nil)
}

// Invoke implements the chaincode shim interface
func (s *GeneralQuerier) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	cid := stub.GetChannelID()
	logger.Infof("Channel ID: %s", cid)

	args := stub.GetArgs()
	if len(args) < 2 {
		return shim.Error(fmt.Sprintf("Incorrect number of arguments, %d", len(args)))
	}
	chaincodeID := string(args[0])
	fname := string(args[1])

	targetLedger := peer.GetLedger(cid)
	if targetLedger == nil {
		return shim.Error(fmt.Sprintf("Invalid chain ID, %s", cid))
	}

	switch fname {
	case GetState:
		if len(args) < 3 {
			return shim.Error(fmt.Sprintf("missing 3rd argument for %s", fname))
		}
		executor, err := targetLedger.NewQueryExecutor()
		if err != nil {
			return shim.Error(fmt.Sprintf("Failed to get QueryExecutor with cid %s, error %s", cid, err))
		}
		return getState(executor, chaincodeID, args[2])
	}

	return shim.Error(fmt.Sprintf("Requested function %s not found.", fname))
}

func getState(executor ledger.QueryExecutor, chaincodeID string, key []byte) pb.Response {
	if key == nil {
		return shim.Error("State key must not be nil")
	}

	value, err := executor.GetState(chaincodeID, string(key))
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get state with namespace %s, error %s", chaincodeID, err))
	}
	return shim.Success(value)
}

func main() {}
