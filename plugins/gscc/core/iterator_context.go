/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package core

import (
	"sync"

	commonledger "github.com/hyperledger/fabric/common/ledger"
)

// IteratorContext ...
type IteratorContext struct {
	mutex               sync.Mutex
	queryIteratorMap    map[string]commonledger.ResultsIterator
	pendingQueryResults map[string]*PendingQueryResult
}

// Initialize ...
func (c *IteratorContext) Initialize(queryID string, iter commonledger.ResultsIterator) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.queryIteratorMap == nil {
		c.queryIteratorMap = make(map[string]commonledger.ResultsIterator)
	}
	if c.pendingQueryResults == nil {
		c.pendingQueryResults = make(map[string]*PendingQueryResult)
	}

	c.queryIteratorMap[queryID] = iter
	c.pendingQueryResults[queryID] = &PendingQueryResult{}
}

// GetQueryIterator ...
func (c *IteratorContext) GetQueryIterator(queryID string) commonledger.ResultsIterator {
	c.mutex.Lock()
	iter := c.queryIteratorMap[queryID]
	c.mutex.Unlock()
	return iter
}

// GetPendingQueryResult ...
func (c *IteratorContext) GetPendingQueryResult(queryID string) *PendingQueryResult {
	c.mutex.Lock()
	result := c.pendingQueryResults[queryID]
	c.mutex.Unlock()
	return result
}

// Cleanup ...
func (c *IteratorContext) Cleanup(queryID string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	iter := c.queryIteratorMap[queryID]
	if iter != nil {
		iter.Close()
	}
	delete(c.queryIteratorMap, queryID)
	delete(c.pendingQueryResults, queryID)
}
