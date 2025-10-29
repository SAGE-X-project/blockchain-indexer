package resolver

import (
	"net/http"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
)

// NewGraphQLHandler creates an HTTP handler for GraphQL requests
func NewGraphQLHandler(resolver *Resolver, enablePlayground bool) http.Handler {
	schema := buildSchema(resolver)

	h := handler.New(&handler.Config{
		Schema:     &schema,
		Pretty:     true,
		GraphiQL:   enablePlayground,
		Playground: enablePlayground,
	})

	return h
}

// buildSchema builds the GraphQL schema from the resolver
func buildSchema(r *Resolver) graphql.Schema {
	// Define custom scalar types
	timestampScalar := graphql.NewScalar(graphql.ScalarConfig{
		Name:        "Time",
		Description: "RFC3339 formatted timestamp",
		Serialize: func(value interface{}) interface{} {
			return value
		},
	})

	bigIntScalar := graphql.NewScalar(graphql.ScalarConfig{
		Name:        "BigInt",
		Description: "Large integer as string",
		Serialize: func(value interface{}) interface{} {
			return value
		},
	})

	// Define Log type
	logType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Log",
		Fields: graphql.Fields{
			"address": &graphql.Field{
				Type: graphql.String,
			},
			"topics": &graphql.Field{
				Type: graphql.NewList(graphql.String),
			},
			"data": &graphql.Field{
				Type: graphql.String,
			},
			"logIndex": &graphql.Field{
				Type: graphql.Int,
			},
		},
	})

	// Define Transaction type
	transactionType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Transaction",
		Fields: graphql.Fields{
			"chainID": &graphql.Field{
				Type: graphql.String,
			},
			"hash": &graphql.Field{
				Type: graphql.String,
			},
			"blockNumber": &graphql.Field{
				Type: bigIntScalar,
			},
			"blockHash": &graphql.Field{
				Type: graphql.String,
			},
			"blockTimestamp": &graphql.Field{
				Type: timestampScalar,
			},
			"txIndex": &graphql.Field{
				Type: graphql.Int,
			},
			"from": &graphql.Field{
				Type: graphql.String,
			},
			"to": &graphql.Field{
				Type: graphql.String,
			},
			"value": &graphql.Field{
				Type: graphql.String,
			},
			"gasPrice": &graphql.Field{
				Type: graphql.String,
			},
			"gasUsed": &graphql.Field{
				Type: bigIntScalar,
			},
			"nonce": &graphql.Field{
				Type: bigIntScalar,
			},
			"status": &graphql.Field{
				Type: graphql.String,
			},
			"input": &graphql.Field{
				Type: graphql.String,
			},
			"contractAddress": &graphql.Field{
				Type: graphql.String,
			},
			"logs": &graphql.Field{
				Type: graphql.NewList(logType),
			},
			"createdAt": &graphql.Field{
				Type: timestampScalar,
			},
		},
	})

	// Define Block type
	blockType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Block",
		Fields: graphql.Fields{
			"chainID": &graphql.Field{
				Type: graphql.String,
			},
			"chainType": &graphql.Field{
				Type: graphql.String,
			},
			"number": &graphql.Field{
				Type: bigIntScalar,
			},
			"hash": &graphql.Field{
				Type: graphql.String,
			},
			"parentHash": &graphql.Field{
				Type: graphql.String,
			},
			"timestamp": &graphql.Field{
				Type: timestampScalar,
			},
			"gasUsed": &graphql.Field{
				Type: bigIntScalar,
			},
			"gasLimit": &graphql.Field{
				Type: bigIntScalar,
			},
			"miner": &graphql.Field{
				Type: graphql.String,
			},
			"txCount": &graphql.Field{
				Type: graphql.Int,
			},
			"transactions": &graphql.Field{
				Type: graphql.NewList(transactionType),
			},
			"createdAt": &graphql.Field{
				Type: timestampScalar,
			},
		},
	})

	// Define Chain type
	chainType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Chain",
		Fields: graphql.Fields{
			"chainID": &graphql.Field{
				Type: graphql.String,
			},
			"chainType": &graphql.Field{
				Type: graphql.String,
			},
			"name": &graphql.Field{
				Type: graphql.String,
			},
			"network": &graphql.Field{
				Type: graphql.String,
			},
			"status": &graphql.Field{
				Type: graphql.String,
			},
			"startBlock": &graphql.Field{
				Type: bigIntScalar,
			},
			"latestIndexedBlock": &graphql.Field{
				Type: bigIntScalar,
			},
			"latestChainBlock": &graphql.Field{
				Type: bigIntScalar,
			},
			"lastUpdated": &graphql.Field{
				Type: timestampScalar,
			},
		},
	})

	// Define Progress type
	progressType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Progress",
		Fields: graphql.Fields{
			"chainID": &graphql.Field{
				Type: graphql.String,
			},
			"chainType": &graphql.Field{
				Type: graphql.String,
			},
			"latestIndexedBlock": &graphql.Field{
				Type: bigIntScalar,
			},
			"latestChainBlock": &graphql.Field{
				Type: bigIntScalar,
			},
			"targetBlock": &graphql.Field{
				Type: bigIntScalar,
			},
			"startBlock": &graphql.Field{
				Type: bigIntScalar,
			},
			"blocksBehind": &graphql.Field{
				Type: bigIntScalar,
			},
			"progressPercentage": &graphql.Field{
				Type: graphql.Float,
			},
			"blocksPerSecond": &graphql.Field{
				Type: graphql.Float,
			},
			"estimatedTimeLeft": &graphql.Field{
				Type: graphql.String,
			},
			"lastUpdated": &graphql.Field{
				Type: timestampScalar,
			},
			"status": &graphql.Field{
				Type: graphql.String,
			},
		},
	})

	// Define Query type
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"block": &graphql.Field{
				Type: blockType,
				Args: graphql.FieldConfigArgument{
					"chainID": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"number": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(bigIntScalar),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					// TODO: Implement resolver call
					return nil, nil
				},
			},
			"transaction": &graphql.Field{
				Type: transactionType,
				Args: graphql.FieldConfigArgument{
					"chainID": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"hash": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					// TODO: Implement resolver call
					return nil, nil
				},
			},
			"chain": &graphql.Field{
				Type: chainType,
				Args: graphql.FieldConfigArgument{
					"chainID": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					// TODO: Implement resolver call
					return nil, nil
				},
			},
			"progress": &graphql.Field{
				Type: progressType,
				Args: graphql.FieldConfigArgument{
					"chainID": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					// TODO: Implement resolver call
					return nil, nil
				},
			},
		},
	})

	schema, _ := graphql.NewSchema(graphql.SchemaConfig{
		Query: queryType,
	})

	return schema
}
