package rediscluster

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/bartventer/gocache/pkg/driver"
	"github.com/bartventer/gocache/pkg/drivertest"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var exposedPorts = []string{
	"7000",
	"7001",
	"7002",
	"7003",
	"7004",
	"7005",
}

func TestRedisClusterCache_OpenCacheURL(t *testing.T) {
	r := &redisClusterCache[string]{}

	u, err := url.Parse("rediscluster://localhost:7000,localhost:7001,localhost:7002,localhost:7003,localhost:7004,localhost:7005?maxretries=5&minretrybackoff=1000ms")
	require.NoError(t, err)

	_, err = r.OpenCacheURL(context.Background(), u)
	require.NoError(t, err)
	assert.NotNil(t, r.client)
}

func TestRedisClusterCache_New(t *testing.T) {
	ctx := context.Background()
	r := New[string](ctx, &Options{
		ClusterOptions: redis.ClusterOptions{
			Addrs: []string{
				"localhost:7000",
				"localhost:7001",
				"localhost:7002",
				"localhost:7003",
				"localhost:7004",
				"localhost:7005",
			},
		},
	})
	require.NotNil(t, r)
	assert.NotNil(t, r.client)
}

// setupRedisCluster creates a new Redis cluster container.
func setupCache[K driver.String](t *testing.T) *redisClusterCache[K] {
	t.Helper()
	// Create a new Redis cluster container
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		// TODO: Switch to latest once this issue is resolved: https://github.com/Grokzen/docker-redis-cluster/issues/162
		Image:        "grokzen/redis-cluster:7.0.10",
		ExposedPorts: exposedPorts,
		ConfigModifier: func(c *container.Config) {
			c.Healthcheck = &container.HealthConfig{
				Test:          []string{"CMD", "redis-cli", "-c", "-p", "7000", "cluster", "info"},
				Interval:      30 * time.Second,
				Timeout:       60 * time.Second,
				Retries:       5,
				StartPeriod:   20 * time.Second,
				StartInterval: 5 * time.Second,
			}
		},
		WaitingFor: wait.ForHealthCheck(),
		Tmpfs:      map[string]string{"/data": "rw"},
	}
	redisClusterC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start Redis cluster container: %v", err)
	}
	t.Cleanup(func() {
		if cleanupErr := redisClusterC.Terminate(ctx); cleanupErr != nil {
			t.Fatalf("Failed to terminate Redis cluster container: %v", cleanupErr)
		}
	})
	// Create a new Redis cluster client
	clusterNodes := make([]string, 0, len(exposedPorts))
	for _, port := range exposedPorts {
		portEndpoint, portErr := redisClusterC.PortEndpoint(ctx, nat.Port(port), "")
		if portErr != nil {
			t.Fatalf("Failed to get port endpoint (%s): %v", port, portErr)
		}
		clusterNodes = append(clusterNodes, portEndpoint)
	}
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:           clusterNodes,
		MaxRetries:      5,
		MinRetryBackoff: 1000 * time.Millisecond,
	})
	t.Cleanup(func() {
		client.Close()
	})
	err = client.ForEachShard(ctx, func(ctx context.Context, client *redis.Client) error {
		return client.Ping(ctx).Err()
	})
	if err != nil {
		t.Fatalf("Failed to ping Redis cluster container: %v", err)
	}
	return &redisClusterCache[K]{client: client, config: &Config{CountLimit: 100}}
}

type harness[K driver.String] struct {
	cache *redisClusterCache[K]
}

func (h *harness[K]) MakeCache(ctx context.Context) (driver.Cache[K], error) {
	return h.cache, nil
}

func (h *harness[K]) Close() {
	// Cleanup is handled in setup function
}

func (h *harness[K]) Options() drivertest.Options {
	return drivertest.Options{
		PatternMatchingDisabled: false,
		CloseIsNoop:             false,
	}
}

func newHarness[K driver.String](ctx context.Context, t *testing.T) (drivertest.Harness[K], error) {
	cache := setupCache[K](t)
	return &harness[K]{
		cache: cache,
	}, nil
}

func TestConformance(t *testing.T) {
	drivertest.RunConformanceTests(t, newHarness[string])
}
