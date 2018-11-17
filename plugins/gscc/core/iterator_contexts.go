/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package core

import (
	"sync"

	commonledger "github.com/hyperledger/fabric/common/ledger"
	"github.com/pkg/errors"
)

// IteratorContexts maintains active history iterator contexts for the gscc.
type IteratorContexts struct {
	mutex    sync.Mutex
	contexts map[string]*IteratorContext
}

// NewIteratorContexts creates a registry for active iterator contexts.
func NewIteratorContexts() *IteratorContexts {
	return &IteratorContexts{
		contexts: map[string]*IteratorContext{},
	}
}

func contextID(chainID, txid string) string {
	return chainID + txid
}

// Create creates a new IteratorContext for the specified chain and
// transaction ID. An error is returned when a iterator context has already
// been created for the specified chain and transaction ID.
func (c *IteratorContexts) Create(chainID, txid string) (*IteratorContext, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	ctxID := contextID(chainID, txid)
	if c.contexts[ctxID] != nil {
		return nil, errors.Errorf("Context: %s(%s) exists", txid, chainID)
	}

	iterctx := &IteratorContext{
		queryIteratorMap: map[string]commonledger.ResultsIterator{},
	}
	c.contexts[ctxID] = iterctx

	return iterctx, nil
}

// Get retrieves the iterator context associated with the chain and
// transaction ID.
func (c *IteratorContexts) Get(chainID, txid string) *IteratorContext {
	ctxID := contextID(chainID, txid)
	c.mutex.Lock()
	ic := c.contexts[ctxID]
	c.mutex.Unlock()
	return ic
}

// Delete removes the iterator context associated with the specified chain
// and transaction ID.
func (c *IteratorContexts) Delete(chainID, txid string) {
	ctxID := contextID(chainID, txid)
	c.mutex.Lock()
	delete(c.contexts, ctxID)
	c.mutex.Unlock()
}
