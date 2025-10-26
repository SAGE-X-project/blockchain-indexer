package pebble

import (
	"fmt"
	"strconv"
	"strings"
)

// Key prefixes for different data types
const (
	// Block data prefixes
	PrefixBlock     = "block:"      // block:{chainID}:{blockNumber}
	PrefixBlockHash = "block_hash:" // block_hash:{chainID}:{hash}

	// Transaction data prefixes
	PrefixTx          = "tx:"        // tx:{chainID}:{txHash}
	PrefixTxByBlock   = "tx_block:"  // tx_block:{chainID}:{blockNumber}:{txIndex}
	PrefixAddrTx      = "addr_tx:"   // addr_tx:{chainID}:{address}:{blockNumber}:{txIndex}

	// Chain configuration prefix
	PrefixChain = "chain:" // chain:{chainID}

	// Metadata prefixes
	PrefixLatestHeight = "latest:"  // latest:{chainID}
	PrefixStats        = "stats:"   // stats:{chainID}

	// Separator for key components
	KeySeparator = ":"
)

// BlockKey generates a key for storing block data by block number
// Format: block:{chainID}:{blockNumber}
func BlockKey(chainID string, blockNumber uint64) []byte {
	return []byte(fmt.Sprintf("%s%s%s%d",
		PrefixBlock, chainID, KeySeparator, blockNumber))
}

// BlockHashKey generates a key for storing block hash to number mapping
// Format: block_hash:{chainID}:{hash}
func BlockHashKey(chainID string, hash string) []byte {
	return []byte(fmt.Sprintf("%s%s%s%s",
		PrefixBlockHash, chainID, KeySeparator, hash))
}

// TransactionKey generates a key for storing transaction data
// Format: tx:{chainID}:{txHash}
func TransactionKey(chainID string, txHash string) []byte {
	return []byte(fmt.Sprintf("%s%s%s%s",
		PrefixTx, chainID, KeySeparator, txHash))
}

// TransactionByBlockKey generates a key for indexing transactions by block
// Format: tx_block:{chainID}:{blockNumber}:{txIndex}
func TransactionByBlockKey(chainID string, blockNumber uint64, txIndex uint64) []byte {
	return []byte(fmt.Sprintf("%s%s%s%d%s%d",
		PrefixTxByBlock, chainID, KeySeparator, blockNumber, KeySeparator, txIndex))
}

// TransactionByBlockPrefix generates a prefix for scanning all transactions in a block
// Format: tx_block:{chainID}:{blockNumber}:
func TransactionByBlockPrefix(chainID string, blockNumber uint64) []byte {
	return []byte(fmt.Sprintf("%s%s%s%d%s",
		PrefixTxByBlock, chainID, KeySeparator, blockNumber, KeySeparator))
}

// AddressTxKey generates a key for indexing transactions by address
// Format: addr_tx:{chainID}:{address}:{blockNumber}:{txIndex}
func AddressTxKey(chainID string, address string, blockNumber uint64, txIndex uint64) []byte {
	return []byte(fmt.Sprintf("%s%s%s%s%s%d%s%d",
		PrefixAddrTx, chainID, KeySeparator, address, KeySeparator, blockNumber, KeySeparator, txIndex))
}

// AddressTxPrefix generates a prefix for scanning all transactions for an address
// Format: addr_tx:{chainID}:{address}:
func AddressTxPrefix(chainID string, address string) []byte {
	return []byte(fmt.Sprintf("%s%s%s%s%s",
		PrefixAddrTx, chainID, KeySeparator, address, KeySeparator))
}

// ChainKey generates a key for storing chain configuration
// Format: chain:{chainID}
func ChainKey(chainID string) []byte {
	return []byte(fmt.Sprintf("%s%s", PrefixChain, chainID))
}

// LatestHeightKey generates a key for storing the latest indexed block height
// Format: latest:{chainID}
func LatestHeightKey(chainID string) []byte {
	return []byte(fmt.Sprintf("%s%s", PrefixLatestHeight, chainID))
}

// StatsKey generates a key for storing chain statistics
// Format: stats:{chainID}
func StatsKey(chainID string) []byte {
	return []byte(fmt.Sprintf("%s%s", PrefixStats, chainID))
}

// ChainStatsKey is an alias for StatsKey for better readability
func ChainStatsKey(chainID string) []byte {
	return StatsKey(chainID)
}

// BlockRangePrefix generates a prefix for scanning blocks in a range
// Format: block:{chainID}:
func BlockRangePrefix(chainID string) []byte {
	return []byte(fmt.Sprintf("%s%s%s", PrefixBlock, chainID, KeySeparator))
}

// ChainPrefix generates a prefix for scanning all chains
// Format: chain:
func ChainPrefix() []byte {
	return []byte(PrefixChain)
}

// ParseBlockKey parses a block key and extracts chainID and blockNumber
func ParseBlockKey(key []byte) (chainID string, blockNumber uint64, err error) {
	keyStr := string(key)
	if !strings.HasPrefix(keyStr, PrefixBlock) {
		return "", 0, fmt.Errorf("invalid block key prefix")
	}

	parts := strings.Split(keyStr[len(PrefixBlock):], KeySeparator)
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid block key format")
	}

	chainID = parts[0]
	blockNumber, err = strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("invalid block number: %w", err)
	}

	return chainID, blockNumber, nil
}

// ParseTransactionByBlockKey parses a transaction-by-block key
func ParseTransactionByBlockKey(key []byte) (chainID string, blockNumber uint64, txIndex uint64, err error) {
	keyStr := string(key)
	if !strings.HasPrefix(keyStr, PrefixTxByBlock) {
		return "", 0, 0, fmt.Errorf("invalid transaction-by-block key prefix")
	}

	parts := strings.Split(keyStr[len(PrefixTxByBlock):], KeySeparator)
	if len(parts) != 3 {
		return "", 0, 0, fmt.Errorf("invalid transaction-by-block key format")
	}

	chainID = parts[0]
	blockNumber, err = strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return "", 0, 0, fmt.Errorf("invalid block number: %w", err)
	}

	txIndex, err = strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		return "", 0, 0, fmt.Errorf("invalid transaction index: %w", err)
	}

	return chainID, blockNumber, txIndex, nil
}

// ParseAddressTxKey parses an address transaction key
func ParseAddressTxKey(key []byte) (chainID string, address string, blockNumber uint64, txIndex uint64, err error) {
	keyStr := string(key)
	if !strings.HasPrefix(keyStr, PrefixAddrTx) {
		return "", "", 0, 0, fmt.Errorf("invalid address-tx key prefix")
	}

	parts := strings.Split(keyStr[len(PrefixAddrTx):], KeySeparator)
	if len(parts) != 4 {
		return "", "", 0, 0, fmt.Errorf("invalid address-tx key format")
	}

	chainID = parts[0]
	address = parts[1]
	blockNumber, err = strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		return "", "", 0, 0, fmt.Errorf("invalid block number: %w", err)
	}

	txIndex, err = strconv.ParseUint(parts[3], 10, 64)
	if err != nil {
		return "", "", 0, 0, fmt.Errorf("invalid transaction index: %w", err)
	}

	return chainID, address, blockNumber, txIndex, nil
}
