package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/torderonex/load-balancer/internal/app"
	"github.com/torderonex/load-balancer/internal/config"
	"github.com/torderonex/load-balancer/internal/handler"
	"github.com/torderonex/load-balancer/internal/storage"
)

// setupTestConfig создает тестовую конфигурацию
func setupTestConfig() *config.Config {
	return &config.Config{
		Logger: config.Logger{
			Level: "debug",
		},
		RateLimiter: config.RateLimiter{
			Enabled:        true,
			DefaultRate:    10,
			Capacity:       10,
			RefillInterval: 5 * time.Second,
		},
		ProxyServer: config.ProxyServer{
			Port:        8080,
			ReadTimeout: 10 * time.Second,
		},
		ApiServer: config.ApiServer{
			Port:        8081,
			ReadTimeout: 10 * time.Second,
		},
		HealthCheck: config.HealthCheck{
			Enabled:       true,
			CheckInterval: 1 * time.Second,
		},
		Balancer: config.Balancer{
			Strategy: "round_robin",
			Backends: []string{},
		},
	}
}

// startTestBackends запускает тестовые бэкенд-серверы
func startTestBackends(t *testing.T, count int) ([]*httptest.Server, []string) {
	var servers []*httptest.Server
	var backendURLs []string

	for i := range count {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("Response from backend %d", i)))
		}))
		servers = append(servers, server)
		backendURLs = append(backendURLs, server.URL)
	}

	return servers, backendURLs
}

var backendTestResponseLen = len("Response from backend d")

func TestBalancer(t *testing.T) {
	servers, backendURLs := startTestBackends(t, 3)
	defer func() {
		for _, server := range servers {
			server.Close()
		}
	}()

	_, server := setupTestApp(t, backendURLs)
	defer server.Close()

	t.Run("Round Robin Distribution", func(t *testing.T) {
		responses := make(map[string]int)
		var mu sync.Mutex

		for range 9 {
			resp, err := http.Get(server.URL)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)

			body := make([]byte, 100)
			n, err := resp.Body.Read(body)
			if err != nil && err != io.EOF {
				require.NoError(t, err)
			}
			bodyStr := string(body[:n])

			mu.Lock()
			responses[bodyStr]++
			mu.Unlock()

			resp.Body.Close()
		}
		for k, v := range responses {
			fmt.Println(k, v)
		}
		require.Len(t, responses, 3, "Requests should be distributed among all backends")

		for _, count := range responses {
			assert.True(t, count >= 2 && count <= 4, "Each backend should receive approximately equal number of requests")
		}
	})

	t.Run("Backend Failure Handling", func(t *testing.T) {
		servers[0].Close()

		time.Sleep(1 * time.Second)

		for range 6 {
			resp, err := http.Get(server.URL)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)

			body := make([]byte, 100)
			n, err := resp.Body.Read(body)
			if err != nil && err != io.EOF {
				require.NoError(t, err)
			}
			bodyStr := string(body[:n])

			assert.NotContains(t, bodyStr, "Response from backend 0", "Request should not be routed to unavailable backend")

			resp.Body.Close()
		}
	})
}

func TestRateLimiter(t *testing.T) {
	servers, backendURLs := startTestBackends(t, 1)
	defer servers[0].Close()

	_, server := setupTestApp(t, backendURLs, withCustomRateLimit(2, 1, 50*time.Second))
	defer server.Close()
	client := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives:   false,
			MaxIdleConnsPerHost: 100,
		},
	}
	t.Run("Rate Limiting Works", func(t *testing.T) {
		for range 2 {
			resp, err := client.Get(server.URL)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}

		resp, err := client.Get(server.URL)
		require.NoError(t, err)
		require.Equal(t, http.StatusTooManyRequests, resp.StatusCode, "Status 429 should be returned when rate limit is exceeded")
		resp.Body.Close()
	})
}

func TestApiServer(t *testing.T) {
	cfg := setupTestConfig()
	storage := storage.New(cfg)
	h := handler.NewHandler(storage)
	router := h.InitRoutes()

	t.Run("Set Client Rate", func(t *testing.T) {
		rateRequest := handler.RateRequest{
			ClientID: "test-client",
			Rate:     50,
		}

		body, _ := json.Marshal(rateRequest)
		req := httptest.NewRequest(http.MethodPost, "/client/rate", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		client, exists := storage.GetClient("test-client")
		require.True(t, exists, "Client should exist in storage")
		require.Equal(t, 50, client.Rate, "Client rate should be updated to 50")
	})

	t.Run("Set Client Capacity", func(t *testing.T) {
		capacityRequest := handler.CapacityRequest{
			ClientID: "test-client",
			Capacity: 100,
		}

		body, _ := json.Marshal(capacityRequest)
		req := httptest.NewRequest(http.MethodPost, "/client/capacity", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		client, exists := storage.GetClient("test-client")
		require.True(t, exists)
		require.Equal(t, 100, client.Capacity)
	})

	t.Run("Invalid Requests", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/client/status", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusNotFound, w.Code)

	})
}

func TestFullSystem(t *testing.T) {

	servers, backendURLs := startTestBackends(t, 3)
	defer func() {
		for _, server := range servers {
			server.Close()
		}
	}()

	cfg := setupTestConfig()
	cfg.Balancer.Backends = backendURLs
	cfg.RateLimiter.DefaultRate = 5
	cfg.RateLimiter.Capacity = 5

	application := app.MustNew(cfg)

	go application.Start()

	time.Sleep(500 * time.Millisecond)

	t.Run("Balancer Distribution", func(t *testing.T) {
		responses := make(map[string]struct{})

		for range 5 {
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d", cfg.ProxyServer.Port))
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)

			body := make([]byte, 100)
			n, _ := resp.Body.Read(body)
			bodyStr := string(body[:n])

			responses[bodyStr] = struct{}{}
			resp.Body.Close()
		}

		require.True(t, len(responses) > 1, "Requests should be distributed between different backends")
	})

	t.Run("API for Client Management", func(t *testing.T) {
		capacityRequest := handler.CapacityRequest{
			ClientID: "::1",
			Capacity: 20,
		}

		body, _ := json.Marshal(capacityRequest)
		resp, err := http.Post(
			fmt.Sprintf("http://localhost:%d/client/capacity", cfg.ApiServer.Port),
			"application/json",
			bytes.NewBuffer(body),
		)

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()

		for range 20 {
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d", cfg.ProxyServer.Port))
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}

		resp, err = http.Get(fmt.Sprintf("http://localhost:%d", cfg.ProxyServer.Port))
		require.NoError(t, err)
		require.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
		resp.Body.Close()
	})
}

// setupTestApp создает тестовое приложение с настроенными компонентами
func setupTestApp(t *testing.T, backendURLs []string, options ...func(*config.Config)) (*app.App, *httptest.Server) {
	cfg := setupTestConfig()
	cfg.Balancer.Backends = backendURLs

	for _, option := range options {
		option(cfg)
	}

	application := app.MustNew(cfg)

	server := httptest.NewServer(application.GetBalancer())

	go application.GetBalancer().StartHealthCheck(100 * time.Millisecond)
	go application.GetLimiter().StartRefill(100 * time.Millisecond)

	return application, server
}

// Option-функции для настройки конфигурации
func withCustomRateLimit(rate int, capacity int, refillInterval time.Duration) func(*config.Config) {
	return func(cfg *config.Config) {
		cfg.RateLimiter.DefaultRate = rate
		cfg.RateLimiter.Capacity = capacity
		cfg.RateLimiter.RefillInterval = refillInterval
	}
}

func withCustomStrategy(strategy string) func(*config.Config) {
	return func(cfg *config.Config) {
		cfg.Balancer.Strategy = strategy
	}
}

func withCustomRefillInterval(interval time.Duration) func(*config.Config) {
	return func(cfg *config.Config) {
		cfg.RateLimiter.RefillInterval = interval
	}
}
