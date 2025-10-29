package cosmos

import (
	"context"
	"fmt"
	"time"

	"github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"
)

// Client wraps the Tendermint/CometBFT RPC client
type Client struct {
	config *Config
	client *http.HTTP
}

// NewClient creates a new Cosmos RPC client
func NewClient(config *Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create HTTP client
	httpClient, err := http.New(config.RPCURL, "/websocket")
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC client: %w", err)
	}

	return &Client{
		config: config,
		client: httpClient,
	}, nil
}

// Start starts the client
func (c *Client) Start() error {
	return c.client.Start()
}

// Stop stops the client
func (c *Client) Stop() error {
	return c.client.Stop()
}

// Status returns the current node status
func (c *Client) Status(ctx context.Context) (*coretypes.ResultStatus, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	return c.client.Status(ctx)
}

// Block returns a block at a given height
func (c *Client) Block(ctx context.Context, height *int64) (*coretypes.ResultBlock, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	return c.client.Block(ctx, height)
}

// BlockByHash returns a block by hash
func (c *Client) BlockByHash(ctx context.Context, hash []byte) (*coretypes.ResultBlock, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	return c.client.BlockByHash(ctx, hash)
}

// BlockResults returns block results at a given height
func (c *Client) BlockResults(ctx context.Context, height *int64) (*coretypes.ResultBlockResults, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	return c.client.BlockResults(ctx, height)
}

// Tx returns a transaction by hash
func (c *Client) Tx(ctx context.Context, hash []byte, prove bool) (*coretypes.ResultTx, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	return c.client.Tx(ctx, hash, prove)
}

// TxSearch searches for transactions
func (c *Client) TxSearch(
	ctx context.Context,
	query string,
	prove bool,
	page, perPage *int,
	orderBy string,
) (*coretypes.ResultTxSearch, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	return c.client.TxSearch(ctx, query, prove, page, perPage, orderBy)
}

// BlockSearch searches for blocks
func (c *Client) BlockSearch(
	ctx context.Context,
	query string,
	page, perPage *int,
	orderBy string,
) (*coretypes.ResultBlockSearch, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	return c.client.BlockSearch(ctx, query, page, perPage, orderBy)
}

// Validators returns validators at a given height
func (c *Client) Validators(
	ctx context.Context,
	height *int64,
	page, perPage *int,
) (*coretypes.ResultValidators, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	return c.client.Validators(ctx, height, page, perPage)
}

// Genesis returns the genesis file
func (c *Client) Genesis(ctx context.Context) (*coretypes.ResultGenesis, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	return c.client.Genesis(ctx)
}

// ABCIInfo returns ABCI info
func (c *Client) ABCIInfo(ctx context.Context) (*coretypes.ResultABCIInfo, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	return c.client.ABCIInfo(ctx)
}

// Subscribe subscribes to events
func (c *Client) Subscribe(
	ctx context.Context,
	subscriber string,
	query string,
	outCapacity ...int,
) (out <-chan coretypes.ResultEvent, err error) {
	return c.client.Subscribe(ctx, subscriber, query, outCapacity...)
}

// Unsubscribe unsubscribes from events
func (c *Client) Unsubscribe(ctx context.Context, subscriber string, query string) error {
	return c.client.Unsubscribe(ctx, subscriber, query)
}

// UnsubscribeAll unsubscribes from all events
func (c *Client) UnsubscribeAll(ctx context.Context, subscriber string) error {
	return c.client.UnsubscribeAll(ctx, subscriber)
}

// Health checks the health of the node
func (c *Client) Health(ctx context.Context) (*coretypes.ResultHealth, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	result, err := c.client.Health(ctx)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetLatestBlockHeight returns the latest block height
func (c *Client) GetLatestBlockHeight(ctx context.Context) (int64, error) {
	status, err := c.Status(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get status: %w", err)
	}

	return status.SyncInfo.LatestBlockHeight, nil
}

// GetBlockByHeight returns a block by height
func (c *Client) GetBlockByHeight(ctx context.Context, height int64) (*tmtypes.Block, error) {
	result, err := c.Block(ctx, &height)
	if err != nil {
		return nil, fmt.Errorf("failed to get block at height %d: %w", height, err)
	}

	return result.Block, nil
}

// GetLatestBlock returns the latest block
func (c *Client) GetLatestBlock(ctx context.Context) (*tmtypes.Block, error) {
	result, err := c.Block(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block: %w", err)
	}

	return result.Block, nil
}

// GetBlockTransactions returns transactions in a block
func (c *Client) GetBlockTransactions(ctx context.Context, height int64) (tmtypes.Txs, error) {
	block, err := c.GetBlockByHeight(ctx, height)
	if err != nil {
		return nil, err
	}

	return block.Txs, nil
}

// withTimeout creates a context with timeout
func (c *Client) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if c.config.Timeout > 0 {
		return context.WithTimeout(ctx, c.config.Timeout)
	}
	return context.WithTimeout(ctx, 30*time.Second)
}
