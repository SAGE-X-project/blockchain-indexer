package evm

import (
	"github.com/ethereum/go-ethereum"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
)

// blockSubscription implements service.BlockSubscription for EVM chains
type blockSubscription struct {
	sub       ethereum.Subscription
	blockChan chan *models.Block
	errChan   chan error
}

// Channel returns the channel that receives new blocks
func (s *blockSubscription) Channel() <-chan *models.Block {
	return s.blockChan
}

// Unsubscribe cancels the subscription
func (s *blockSubscription) Unsubscribe() {
	if s.sub != nil {
		s.sub.Unsubscribe()
	}
}

// Err returns any subscription error
func (s *blockSubscription) Err() <-chan error {
	return s.errChan
}

// transactionSubscription implements service.TransactionSubscription for EVM chains
type transactionSubscription struct {
	sub   ethereum.Subscription
	txChan chan *models.Transaction
	errChan chan error
}

// Channel returns the channel that receives new transactions
func (s *transactionSubscription) Channel() <-chan *models.Transaction {
	return s.txChan
}

// Unsubscribe cancels the subscription
func (s *transactionSubscription) Unsubscribe() {
	if s.sub != nil {
		s.sub.Unsubscribe()
	}
}

// Err returns any subscription error
func (s *transactionSubscription) Err() <-chan error {
	return s.errChan
}
