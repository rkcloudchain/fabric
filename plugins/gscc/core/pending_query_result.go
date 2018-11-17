/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package core

import (
	"github.com/golang/protobuf/proto"
	commonledger "github.com/hyperledger/fabric/common/ledger"
	"github.com/hyperledger/fabric/plugins/gscc/protos"
)

// PendingQueryResult ...
type PendingQueryResult struct {
	batch []*protos.QueryResultBytes
}

// Cut ...
func (p *PendingQueryResult) Cut() []*protos.QueryResultBytes {
	batch := p.batch
	p.batch = nil
	return batch
}

// Add ...
func (p *PendingQueryResult) Add(queryResult commonledger.QueryResult) error {
	queryResultBytes, err := proto.Marshal(queryResult.(proto.Message))
	if err != nil {
		logger.Errorf("failed to marshal query result: %s", err)
		return err
	}
	p.batch = append(p.batch, &protos.QueryResultBytes{ResultBytes: queryResultBytes})
	return nil
}

// Size ...
func (p *PendingQueryResult) Size() int {
	return len(p.batch)
}
