/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/core/chaincode"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/ledger"
	"github.com/hyperledger/fabric/core/ledger/ledgerconfig"
	"github.com/hyperledger/fabric/core/peer"
	"github.com/hyperledger/fabric/plugins/gscc/core"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
)

// New returns an implementation of the chaincode interface
func New() shim.Chaincode {
	return &GeneralQuerier{
		UUIDGenerator:        chaincode.UUIDGeneratorFunc(util.GenerateUUID),
		IteratorContexts:     core.NewIteratorContexts(),
		QueryResponseBuilder: &core.QueryResponseGenerator{MaxResultLimit: 100},
	}
}

// GeneralQuerier implements cross-chaincode ledger query
type GeneralQuerier struct {
	UUIDGenerator        chaincode.UUIDGenerator
	IteratorContexts     *core.IteratorContexts
	QueryResponseBuilder *core.QueryResponseGenerator
}

var logger = flogging.MustGetLogger("gscc")

// These are function names from Invoke first parameter
const (
	GetState         string = "GetState"
	GetHistoryForKey string = "GetHistoryForKey"
	QueryStateClose  string = "QueryStateClose"
	QueryStateNext   string = "QueryStateNext"
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
	fname := string(args[0])

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
		defer executor.Done()

		return s.handleGetState(executor, args[1], args[2])

	case GetHistoryForKey:
		if len(args) < 3 {
			return shim.Error(fmt.Sprintf("missing 3rd argument for %s", fname))
		}
		historyExecutor, err := targetLedger.NewHistoryQueryExecutor()
		if err != nil {
			return shim.Error(fmt.Sprintf("Failed to get QueryExecutor with cid %s, error %s", cid, err))
		}
		return s.handleGetHistoryForKey(historyExecutor, cid, stub.GetTxID(), args[1], args[2])

	case QueryStateClose:
		if len(args) < 3 {
			return shim.Error(fmt.Sprintf("missing 3rd argument for %s", fname))
		}
		return s.handleQueryStateClose(cid, args[1], args[2])

	case QueryStateNext:
		if len(args) < 3 {
			return shim.Error(fmt.Sprintf("missing 3rd argument for %s", fname))
		}
		return s.handleQueryStateNext(cid, args[1], args[2])
	}

	return shim.Error(fmt.Sprintf("Requested function %s not found.", fname))
}

// Handles query to ledger to get state
func (s *GeneralQuerier) handleGetState(executor ledger.QueryExecutor, ccid, key []byte) pb.Response {
	if key == nil {
		return shim.Error("State key must not be nil")
	}

	value, err := executor.GetState(string(ccid), string(key))
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get state with namespace %s, error %s", ccid, err))
	}
	return shim.Success(value)
}

// Handles query to ledger history db
func (s *GeneralQuerier) handleGetHistoryForKey(executor ledger.HistoryQueryExecutor, cid, tid string, ccid, key []byte) pb.Response {
	iterctx, err := s.IteratorContexts.Create(cid, tid)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to create IteratorContext with %s(%s), error: %s", cid, tid, err))
	}

	iterID := s.UUIDGenerator.New()
	historyIter, err := executor.GetHistoryForKey(string(ccid), string(key))
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get history iterator with namespace %s, error %s", ccid, err))
	}

	totalReturnLimit := calculateTotalReturnLimit()
	iterctx.Initialize(iterID, historyIter)

	payload, err := s.QueryResponseBuilder.Build(iterctx, historyIter, tid, iterID, totalReturnLimit)
	if err != nil {
		iterctx.Cleanup(iterID)
		return shim.Error(fmt.Sprintf("Failed to get payload with %s(%s/%s), error %s", iterID, cid, tid, err))
	}

	payloadBytes, err := proto.Marshal(payload)
	if err != nil {
		iterctx.Cleanup(iterID)
		return shim.Error(fmt.Sprintf("marshal failed: %s", err))
	}

	return shim.Success(payloadBytes)
}

// Handles query to ledger for query state next
func (s *GeneralQuerier) handleQueryStateNext(cid string, txid, iterID []byte) pb.Response {
	iterctx := s.IteratorContexts.Get(cid, string(txid))
	if iterctx == nil {
		err := errors.Errorf("no ledger context: %s %s\n\n", cid, txid)
		return shim.Error(err.Error())
	}

	queryIter := iterctx.GetQueryIterator(string(iterID))
	if queryIter == nil {
		return shim.Error("query iterator not found")
	}

	totalReturnLimit := calculateTotalReturnLimit()
	payload, err := s.QueryResponseBuilder.Build(iterctx, queryIter, string(txid), string(iterID), totalReturnLimit)
	if err != nil {
		iterctx.Cleanup(string(iterID))
		return shim.Error(fmt.Sprintf("Failed to get payload with %s(%s/%s), error %s", string(iterID), cid, string(txid), err))
	}

	payloadBytes, err := proto.Marshal(payload)
	if err != nil {
		iterctx.Cleanup(string(iterID))
		return shim.Error(fmt.Sprintf("marshal failed: %s", err))
	}

	return shim.Success(payloadBytes)
}

// Handles the closing of a state iterator
func (s *GeneralQuerier) handleQueryStateClose(cid string, txid, iterID []byte) pb.Response {
	iterctx := s.IteratorContexts.Get(cid, string(txid))
	if iterctx != nil {
		iterctx.Cleanup(string(iterID))
	} else {
		logger.Warnf("Can't find any IteratorContext with %s(%s)", cid, txid)
	}
	s.IteratorContexts.Delete(cid, string(txid))

	return shim.Success(nil)
}

func calculateTotalReturnLimit() int32 {
	return int32(ledgerconfig.GetTotalQueryLimit())
}

func main() {}
