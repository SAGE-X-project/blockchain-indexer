package polkadot

import (
	"context"
	"fmt"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

// Client wraps the Substrate RPC client
type Client struct {
	config *Config
	api    *gsrpc.SubstrateAPI
}

// NewClient creates a new Polkadot RPC client
func NewClient(config *Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create API client
	api, err := gsrpc.NewSubstrateAPI(config.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create Substrate API: %w", err)
	}

	return &Client{
		config: config,
		api:    api,
	}, nil
}

// Close closes the client connection
func (c *Client) Close() error {
	// Substrate RPC client doesn't require explicit close
	return nil
}

// GetLatestBlockNumber returns the latest finalized block number
func (c *Client) GetLatestBlockNumber(ctx context.Context) (uint64, error) {
	hash, err := c.api.RPC.Chain.GetFinalizedHead()
	if err != nil {
		return 0, fmt.Errorf("failed to get finalized head: %w", err)
	}

	header, err := c.api.RPC.Chain.GetHeader(hash)
	if err != nil {
		return 0, fmt.Errorf("failed to get header: %w", err)
	}

	return uint64(header.Number), nil
}

// GetBlockHash returns the block hash for a given block number
func (c *Client) GetBlockHash(ctx context.Context, blockNumber uint64) (types.Hash, error) {
	hash, err := c.api.RPC.Chain.GetBlockHash(blockNumber)
	if err != nil {
		return types.Hash{}, fmt.Errorf("failed to get block hash for block %d: %w", blockNumber, err)
	}

	return hash, nil
}

// GetBlock returns a block by hash
func (c *Client) GetBlock(ctx context.Context, hash types.Hash) (*types.SignedBlock, error) {
	block, err := c.api.RPC.Chain.GetBlock(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get block %s: %w", hash.Hex(), err)
	}

	return block, nil
}

// GetBlockByNumber returns a block by number
func (c *Client) GetBlockByNumber(ctx context.Context, blockNumber uint64) (*types.SignedBlock, error) {
	hash, err := c.GetBlockHash(ctx, blockNumber)
	if err != nil {
		return nil, err
	}

	return c.GetBlock(ctx, hash)
}

// GetHeader returns a block header by hash
func (c *Client) GetHeader(ctx context.Context, hash types.Hash) (*types.Header, error) {
	header, err := c.api.RPC.Chain.GetHeader(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get header %s: %w", hash.Hex(), err)
	}

	return header, nil
}

// GetHeaderByNumber returns a block header by number
func (c *Client) GetHeaderByNumber(ctx context.Context, blockNumber uint64) (*types.Header, error) {
	hash, err := c.GetBlockHash(ctx, blockNumber)
	if err != nil {
		return nil, err
	}

	return c.GetHeader(ctx, hash)
}

// GetMetadata returns the chain metadata
func (c *Client) GetMetadata(ctx context.Context, blockHash types.Hash) (*types.Metadata, error) {
	meta, err := c.api.RPC.State.GetMetadata(blockHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	return meta, nil
}

// GetRuntimeVersion returns the runtime version
func (c *Client) GetRuntimeVersion(ctx context.Context, blockHash types.Hash) (*types.RuntimeVersion, error) {
	rv, err := c.api.RPC.State.GetRuntimeVersion(blockHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get runtime version: %w", err)
	}

	return rv, nil
}

// GetStorageRaw returns raw storage value
func (c *Client) GetStorageRaw(ctx context.Context, key types.StorageKey, blockHash types.Hash) (*types.StorageDataRaw, error) {
	data, err := c.api.RPC.State.GetStorageRaw(key, blockHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage: %w", err)
	}

	return data, nil
}

// QueryStorage queries storage changes
func (c *Client) QueryStorage(ctx context.Context, keys []types.StorageKey, startBlock, endBlock types.Hash) ([]types.StorageChangeSet, error) {
	changes, err := c.api.RPC.State.QueryStorage(keys, startBlock, endBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to query storage: %w", err)
	}

	return changes, nil
}

// GetSystemHealth returns the system health
func (c *Client) GetSystemHealth(ctx context.Context) (*types.Health, error) {
	health, err := c.api.RPC.System.Health()
	if err != nil {
		return nil, fmt.Errorf("failed to get system health: %w", err)
	}

	return &health, nil
}

// GetSystemChain returns the chain name
func (c *Client) GetSystemChain(ctx context.Context) (string, error) {
	chain, err := c.api.RPC.System.Chain()
	if err != nil {
		return "", fmt.Errorf("failed to get system chain: %w", err)
	}

	return string(chain), nil
}

// GetSystemName returns the node name
func (c *Client) GetSystemName(ctx context.Context) (string, error) {
	name, err := c.api.RPC.System.Name()
	if err != nil {
		return "", fmt.Errorf("failed to get system name: %w", err)
	}

	return string(name), nil
}

// GetSystemVersion returns the node version
func (c *Client) GetSystemVersion(ctx context.Context) (string, error) {
	version, err := c.api.RPC.System.Version()
	if err != nil {
		return "", fmt.Errorf("failed to get system version: %w", err)
	}

	return string(version), nil
}

// GetSystemProperties returns the system properties
func (c *Client) GetSystemProperties(ctx context.Context) (types.ChainProperties, error) {
	props, err := c.api.RPC.System.Properties()
	if err != nil {
		return types.ChainProperties{}, fmt.Errorf("failed to get system properties: %w", err)
	}

	return props, nil
}

// GetAccountInfo returns account information
func (c *Client) GetAccountInfo(ctx context.Context, accountID types.AccountID, blockHash types.Hash) (*types.AccountInfo, error) {
	meta, err := c.GetMetadata(ctx, blockHash)
	if err != nil {
		return nil, err
	}

	key, err := types.CreateStorageKey(meta, "System", "Account", accountID[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create storage key: %w", err)
	}

	var accountInfo types.AccountInfo
	ok, err := c.api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to get account info: %w", err)
	}

	if !ok {
		return nil, fmt.Errorf("account not found")
	}

	return &accountInfo, nil
}

// SubscribeNewHeads subscribes to new block headers
// Note: WebSocket subscriptions require additional setup
func (c *Client) SubscribeNewHeads(ctx context.Context) error {
	if !c.config.EnableWebSocket {
		return fmt.Errorf("websocket is not enabled")
	}

	// Subscription functionality requires WebSocket setup
	// This is a placeholder for future implementation
	return fmt.Errorf("subscription not yet implemented")
}

// SubscribeFinalizedHeads subscribes to finalized block headers
// Note: WebSocket subscriptions require additional setup
func (c *Client) SubscribeFinalizedHeads(ctx context.Context) error {
	if !c.config.EnableWebSocket {
		return fmt.Errorf("websocket is not enabled")
	}

	// Subscription functionality requires WebSocket setup
	// This is a placeholder for future implementation
	return fmt.Errorf("subscription not yet implemented")
}

// GetAPI returns the underlying Substrate API
func (c *Client) GetAPI() *gsrpc.SubstrateAPI {
	return c.api
}
