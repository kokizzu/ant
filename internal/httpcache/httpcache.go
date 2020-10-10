// Package httpcache implements an HTTP client that caches responses.
package httpcache

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// Freshness enumerates freshness.
type Freshness int

// String implementation.
func (f Freshness) String() string {
	switch f {
	case Fresh:
		return "fresh"
	case Stale:
		return "stale"
	case Transparent:
		return "transprent"
	default:
		return fmt.Sprintf("httpcache.Freshness(%d)", f)
	}
}

// All Freshness types.
const (
	Fresh Freshness = iota
	Stale
	Transparent
)

// Strategy represents a cache strategy.
type Strategy interface {
	// Cache returns true if the request is cacheable.
	//
	// The method is called just before a the storage lookup
	// is made.
	cache(req *http.Request) bool

	// Store returns true if the response can be stored.
	//
	// The method is called just before a response is stored.
	store(resp *http.Response) bool

	// Fresh returns true if the response is fresh.
	//
	// The method is called just before a cached response
	// is returned from the cache.
	fresh(resp *http.Response) Freshness
}

// Storage represents the cache storage.
//
// A storage must be safe to use from multiple goroutines.
type Storage interface {
	// Store stores the given response.
	//
	// The method is called just after the response's body is
	// closed, the value contains the full response including headers.
	Store(ctx context.Context, key uint64, value []byte) error

	// Load loads a response by its key.
	//
	// When the response is not found, the method returns a nil
	// byteslice and a nil error.
	//
	// The method returns the full response, as stored by `Store()`.
	Load(ctx context.Context, key uint64) ([]byte, error)
}

// Client represents an HTTP client.
type Client interface {
	// Do performs the given request.
	Do(req *http.Request) (*http.Response, error)
}

// Option represents a cache option.
type Option func(*Cache) error

// WithStrategy sets the strategy to s.
//
// If nil, NewCache returns an error.
func WithStrategy(s Strategy) Option {
	return func(c *Cache) error {
		if s == nil {
			return errors.New("httpcache: strategy must be non-nil")
		}
		c.strategy = s
		return nil
	}
}

// WithStorage sets the storage to s.
//
// If nil, NewCache returns an error.
func WithStorage(s Storage) Option {
	return func(c *Cache) error {
		if s == nil {
			return errors.New("httpcache: storage must be non-nil")
		}
		c.storage = s
		return nil
	}
}

// WithClient sets the HTTP client.
//
// If nil, http.DefaultClient is used.
func WithClient(client Client) Option {
	return func(c *Cache) error {
		if client == nil {
			return errors.New("httpcache: client must be non-nil")
		}
		c.client = client
		return nil
	}
}

// Cache implements an HTTP cache.
type Cache struct {
	storage  Storage
	strategy Strategy
	client   Client
}

// NewCache returns a new cache with the given options.
func NewCache(opts ...Option) (*Cache, error) {
	var cache = &Cache{
		strategy: RFC7234{},
		storage:  &Memstore{},
		client:   http.DefaultClient,
	}

	for _, opt := range opts {
		if err := opt(cache); err != nil {
			return nil, err
		}
	}

	return cache, nil
}
