package health

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/repository"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/logger"
	"go.uber.org/zap"
)

// Status represents the health status
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
)

// CheckResult represents the result of a health check
type CheckResult struct {
	Name      string                 `json:"name"`
	Status    Status                 `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  time.Duration          `json:"duration"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// HealthReport represents the overall health status
type HealthReport struct {
	Status    Status                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]CheckResult `json:"checks"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Check is a function that performs a health check
type Check func(ctx context.Context) CheckResult

// Checker manages health checks
type Checker struct {
	checks     map[string]Check
	mu         sync.RWMutex
	logger     *logger.Logger
	interval   time.Duration
	lastReport *HealthReport
}

// NewChecker creates a new health checker
func NewChecker(logger *logger.Logger, interval time.Duration) *Checker {
	return &Checker{
		checks:   make(map[string]Check),
		logger:   logger,
		interval: interval,
	}
}

// RegisterCheck adds a health check
func (c *Checker) RegisterCheck(name string, check Check) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.checks[name] = check
}

// RunChecks executes all registered health checks
func (c *Checker) RunChecks(ctx context.Context) *HealthReport {
	c.mu.RLock()
	checks := make(map[string]Check, len(c.checks))
	for name, check := range c.checks {
		checks[name] = check
	}
	c.mu.RUnlock()

	results := make(map[string]CheckResult)
	var wg sync.WaitGroup

	// Run checks concurrently
	resultsCh := make(chan CheckResult, len(checks))
	for name, check := range checks {
		wg.Add(1)
		go func(name string, check Check) {
			defer wg.Done()
			result := check(ctx)
			result.Name = name
			resultsCh <- result
		}(name, check)
	}

	// Close channel when all checks complete
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Collect results
	for result := range resultsCh {
		results[result.Name] = result
	}

	// Determine overall status
	overallStatus := StatusHealthy
	for _, result := range results {
		if result.Status == StatusUnhealthy {
			overallStatus = StatusUnhealthy
			break
		}
		if result.Status == StatusDegraded && overallStatus == StatusHealthy {
			overallStatus = StatusDegraded
		}
	}

	report := &HealthReport{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Checks:    results,
	}

	c.mu.Lock()
	c.lastReport = report
	c.mu.Unlock()

	return report
}

// GetLastReport returns the most recent health report
func (c *Checker) GetLastReport() *HealthReport {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastReport
}

// StartPeriodicChecks starts background health checking
func (c *Checker) StartPeriodicChecks(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	// Run initial check
	c.RunChecks(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			report := c.RunChecks(ctx)
			if report.Status != StatusHealthy {
				c.logger.Warn("health check detected issues",
					zap.String("status", string(report.Status)),
					zap.Int("checks", len(report.Checks)),
				)
			}
		}
	}
}

// Common health checks

// StorageHealthCheck checks database connectivity and health
func StorageHealthCheck(chainRepo repository.ChainRepository) Check {
	return func(ctx context.Context) CheckResult {
		start := time.Now()
		result := CheckResult{
			Timestamp: start,
			Details:   make(map[string]interface{}),
		}

		// Try to query chains
		chains, err := chainRepo.GetAllChains(ctx)
		duration := time.Since(start)
		result.Duration = duration

		if err != nil {
			result.Status = StatusUnhealthy
			result.Message = fmt.Sprintf("Storage query failed: %v", err)
			return result
		}

		result.Details["chain_count"] = len(chains)
		result.Details["query_duration_ms"] = duration.Milliseconds()

		// Check query latency
		if duration > 1*time.Second {
			result.Status = StatusDegraded
			result.Message = "Storage queries are slow"
		} else {
			result.Status = StatusHealthy
			result.Message = "Storage is healthy"
		}

		return result
	}
}

// MemoryHealthCheck checks memory usage
func MemoryHealthCheck(thresholdMB uint64) Check {
	return func(ctx context.Context) CheckResult {
		start := time.Now()
		result := CheckResult{
			Timestamp: start,
			Details:   make(map[string]interface{}),
		}

		// Get memory stats
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		duration := time.Since(start)
		result.Duration = duration

		allocMB := m.Alloc / 1024 / 1024
		result.Details["alloc_mb"] = allocMB
		result.Details["sys_mb"] = m.Sys / 1024 / 1024
		result.Details["num_gc"] = m.NumGC

		if allocMB > thresholdMB {
			result.Status = StatusDegraded
			result.Message = fmt.Sprintf("High memory usage: %dMB (threshold: %dMB)", allocMB, thresholdMB)
		} else {
			result.Status = StatusHealthy
			result.Message = fmt.Sprintf("Memory usage: %dMB", allocMB)
		}

		return result
	}
}

// GoroutineHealthCheck checks goroutine count
func GoroutineHealthCheck(threshold int) Check {
	return func(ctx context.Context) CheckResult {
		start := time.Now()
		result := CheckResult{
			Timestamp: start,
			Details:   make(map[string]interface{}),
		}

		count := runtime.NumGoroutine()
		result.Duration = time.Since(start)
		result.Details["goroutine_count"] = count

		if count > threshold {
			result.Status = StatusDegraded
			result.Message = fmt.Sprintf("High goroutine count: %d (threshold: %d)", count, threshold)
		} else {
			result.Status = StatusHealthy
			result.Message = fmt.Sprintf("Goroutine count: %d", count)
		}

		return result
	}
}

// ChainConnectivityCheck checks if chain adapters are reachable
func ChainConnectivityCheck(chainID string, checkFunc func(context.Context) error) Check {
	return func(ctx context.Context) CheckResult {
		start := time.Now()
		result := CheckResult{
			Timestamp: start,
			Details:   make(map[string]interface{}),
		}

		result.Details["chain_id"] = chainID

		err := checkFunc(ctx)
		duration := time.Since(start)
		result.Duration = duration

		if err != nil {
			result.Status = StatusUnhealthy
			result.Message = fmt.Sprintf("Chain %s unreachable: %v", chainID, err)
		} else {
			result.Status = StatusHealthy
			result.Message = fmt.Sprintf("Chain %s is reachable", chainID)
		}

		return result
	}
}
