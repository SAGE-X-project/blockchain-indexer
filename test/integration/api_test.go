package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	indexerv1 "github.com/sage-x-project/blockchain-indexer/api/proto/indexer/v1"
)

const (
	httpAddr = "http://localhost:8080"
	grpcAddr = "localhost:50051"
)

// TestHealthEndpoint tests the health check endpoint
func TestHealthEndpoint(t *testing.T) {
	resp, err := http.Get(httpAddr + "/health")
	if err != nil {
		t.Fatalf("Failed to call health endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if status, ok := result["status"].(string); !ok || status != "ok" {
		t.Errorf("Expected status 'ok', got %v", result["status"])
	}
}

// TestRESTAPIRoot tests the REST API root endpoint
func TestRESTAPIRoot(t *testing.T) {
	resp, err := http.Get(httpAddr + "/api/")
	if err != nil {
		t.Fatalf("Failed to call REST API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if message, ok := result["message"].(string); !ok || message == "" {
		t.Errorf("Expected non-empty message, got %v", result["message"])
	}
}

// TestGraphQLIntrospection tests GraphQL introspection query
func TestGraphQLIntrospection(t *testing.T) {
	query := `{"query":"{ __schema { types { name } } }"}`

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("POST", httpAddr+"/graphql", bytes.NewBufferString(query))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to call GraphQL endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if data, ok := result["data"].(map[string]interface{}); !ok || data == nil {
		t.Errorf("Expected data object, got %v", result)
	}
}

// TestGRPCConnection tests gRPC server connectivity
func TestGRPCConnection(t *testing.T) {
	// Create gRPC connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	// Create client
	client := indexerv1.NewIndexerServiceClient(conn)

	// Test ListChains (should return empty or error, but connection should work)
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()

	resp, err := client.ListChains(ctx2, &indexerv1.ListChainsRequest{})
	if err != nil {
		// It's okay if there are no chains yet, we're just testing connectivity
		t.Logf("ListChains returned error (expected if no chains indexed): %v", err)
	} else {
		t.Logf("ListChains succeeded, found %d chains", len(resp.Chains))
	}
}

// TestMetricsEndpoint tests the Prometheus metrics endpoint
func TestMetricsEndpoint(t *testing.T) {
	resp, err := http.Get("http://localhost:9091/metrics")
	if err != nil {
		t.Fatalf("Failed to call metrics endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Read the entire body
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	// Metrics endpoint is accessible and returns 200 OK
	// Content may be empty initially if no metrics have been registered
	t.Logf("Metrics endpoint returned %d bytes (may be empty initially)", buf.Len())
}

// TestServerEndpoints verifies all server endpoints are accessible
func TestServerEndpoints(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		wantCode int
	}{
		{"Health Check", httpAddr + "/health", http.StatusOK},
		{"REST API Root", httpAddr + "/api/", http.StatusOK},
		{"GraphQL Endpoint", httpAddr + "/graphql", http.StatusOK},
		{"Metrics", "http://localhost:9091/metrics", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(tt.url)
			if err != nil {
				t.Fatalf("Failed to call %s: %v", tt.url, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantCode {
				t.Errorf("Expected status %d, got %d for %s", tt.wantCode, resp.StatusCode, tt.url)
			} else {
				fmt.Printf("✓ %s is accessible\n", tt.name)
			}
		})
	}
}
