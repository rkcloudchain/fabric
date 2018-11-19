/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package core

import (
	"github.com/hyperledger/fabric/common/flogging"
	commonledger "github.com/hyperledger/fabric/common/ledger"
	"github.com/hyperledger/fabric/plugins/gscc/protos"
)

var logger = flogging.MustGetLogger("gscc/core")

// QueryResponseGenerator ...
type QueryResponseGenerator struct {
	MaxResultLimit int
}

// Build takes an iterator and fetch state to construct QueryResponse
func (q *QueryResponseGenerator) Build(iterContext *IteratorContext, iter commonledger.ResultsIterator,
	txid, iterID string, totalReturnLimit int32) (*protos.RangeQueryResponse, error) {

	pendingQueryResults := iterContext.GetPendingQueryResult(iterID)
	totalReturnCount := iterContext.GetTotalReturnCount(iterID)

	for {
		if *totalReturnCount >= totalReturnLimit {
			return createQueryResponse(iterContext, txid, iterID, pendingQueryResults), nil
		}

		queryResult, err := iter.Next()
		switch {
		case err != nil:
			logger.Errorf("Failed to get query result from iterator: %s", iterID)
			iterContext.Cleanup(iterID)
			return nil, err

		case queryResult == nil:
			return createQueryResponse(iterContext, txid, iterID, pendingQueryResults), nil

		case pendingQueryResults.Size() == q.MaxResultLimit:
			batch := pendingQueryResults.Cut()
			if err := pendingQueryResults.Add(queryResult); err != nil {
				iterContext.Cleanup(iterID)
				return nil, err
			}
			*totalReturnCount++
			return &protos.RangeQueryResponse{Results: batch, HasMore: true, Id: iterID, TxId: txid}, nil

		default:
			if err := pendingQueryResults.Add(queryResult); err != nil {
				iterContext.Cleanup(iterID)
				return nil, err
			}
			*totalReturnCount++
		}
	}
}

func createQueryResponse(iterContext *IteratorContext, txid, iterID string, pendingQueryResults *PendingQueryResult) *protos.RangeQueryResponse {
	batch := pendingQueryResults.Cut()
	iterContext.Cleanup(iterID)
	return &protos.RangeQueryResponse{Results: batch, HasMore: false, Id: iterID, TxId: txid}
}
