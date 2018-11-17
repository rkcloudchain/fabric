/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/core/chaincode/shim"
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
	return shim.Success(nil)
}

func main() {}
