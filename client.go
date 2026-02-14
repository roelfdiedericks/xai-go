package xai

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"time"

	v1 "github.com/roelfdiedericks/xai-go/proto/xai/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

const (
	// DefaultEndpoint is the default xAI API endpoint.
	DefaultEndpoint = "api.x.ai:443"
	// DefaultTimeout is the default request timeout.
	DefaultTimeout = 120 * time.Second
	// DefaultModel is the default chat model to use.
	DefaultModel = "grok-4-1-fast-reasoning"
	// DefaultImageModel is the default image generation model to use.
	DefaultImageModel = "grok-2-image"
	// EnvAPIKey is the environment variable for the API key.
	EnvAPIKey = "XAI_APIKEY"
	// DefaultKeepaliveTime is how often to send keepalive pings.
	DefaultKeepaliveTime = 30 * time.Second
	// DefaultKeepaliveTimeout is how long to wait for a keepalive response.
	DefaultKeepaliveTimeout = 10 * time.Second
)

// Config holds the configuration for an xAI client.
type Config struct {
	// Endpoint is the gRPC endpoint (default: api.x.ai:443).
	Endpoint string
	// APIKey is the xAI API key (required).
	APIKey *SecureString
	// Timeout is the default request timeout (default: 120s).
	Timeout time.Duration
	// DefaultModel is the model to use when not specified.
	DefaultModel string
	// TLSConfig allows custom TLS configuration. If nil, uses default TLS.
	TLSConfig *tls.Config
	// KeepaliveTime is how often to send keepalive pings (default: 30s).
	// Set to 0 to use the default, or -1 to disable keepalive.
	KeepaliveTime time.Duration
	// KeepaliveTimeout is how long to wait for a keepalive response (default: 10s).
	KeepaliveTimeout time.Duration
	// KeepalivePermitWithoutStream allows pings when no active streams (default: true).
	// Set to false to only ping during active requests.
	KeepalivePermitWithoutStream *bool
}

// validate checks the config and sets defaults.
func (c *Config) validate() error {
	if c.APIKey == nil || c.APIKey.IsZero() {
		return &Error{
			Code:    ErrAuth,
			Message: "API key is required",
		}
	}
	if c.Endpoint == "" {
		c.Endpoint = DefaultEndpoint
	}
	if c.Timeout == 0 {
		c.Timeout = DefaultTimeout
	}
	if c.DefaultModel == "" {
		c.DefaultModel = DefaultModel
	}
	if c.KeepaliveTime == 0 {
		c.KeepaliveTime = DefaultKeepaliveTime
	}
	if c.KeepaliveTimeout == 0 {
		c.KeepaliveTimeout = DefaultKeepaliveTimeout
	}
	return nil
}

// Client is the xAI API client.
type Client struct {
	conn   *grpc.ClientConn
	config Config

	// Service clients
	chat      v1.ChatClient
	models    v1.ModelsClient
	embedder  v1.EmbedderClient
	tokenizer v1.TokenizeClient
	auth      v1.AuthClient
	sampler   v1.SampleClient
	image     v1.ImageClient
	documents v1.DocumentsClient
	batch     v1.BatchMgmtClient
}

// New creates a new xAI client with the given configuration.
func New(cfg Config) (*Client, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	// Build gRPC dial options
	opts := []grpc.DialOption{
		grpc.WithPerRPCCredentials(&bearerAuth{apiKey: cfg.APIKey}),
	}

	// Add keepalive if not disabled (KeepaliveTime == -1 disables)
	if cfg.KeepaliveTime >= 0 {
		permitWithoutStream := true
		if cfg.KeepalivePermitWithoutStream != nil {
			permitWithoutStream = *cfg.KeepalivePermitWithoutStream
		}
		opts = append(opts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                cfg.KeepaliveTime,
			Timeout:             cfg.KeepaliveTimeout,
			PermitWithoutStream: permitWithoutStream,
		}))
	}

	// Configure TLS
	var creds credentials.TransportCredentials
	if cfg.TLSConfig != nil {
		creds = credentials.NewTLS(cfg.TLSConfig)
	} else {
		creds = credentials.NewTLS(&tls.Config{
			MinVersion: tls.VersionTLS12,
		})
	}
	opts = append(opts, grpc.WithTransportCredentials(creds))

	// Connect
	conn, err := grpc.NewClient(cfg.Endpoint, opts...)
	if err != nil {
		return nil, &Error{
			Code:    ErrUnavailable,
			Message: fmt.Sprintf("failed to connect to %s", cfg.Endpoint),
			Cause:   err,
		}
	}

	return newClientFromConn(conn, cfg), nil
}

// FromEnv creates a new client using the XAI_APIKEY environment variable.
func FromEnv() (*Client, error) {
	apiKey := os.Getenv(EnvAPIKey)
	if apiKey == "" {
		return nil, &Error{
			Code:    ErrAuth,
			Message: fmt.Sprintf("environment variable %s is not set", EnvAPIKey),
		}
	}
	return New(Config{
		APIKey: NewSecureString(apiKey),
	})
}

// WithChannel creates a client using an existing gRPC connection.
// This is useful for custom TLS configurations or connection pooling.
func WithChannel(conn *grpc.ClientConn, apiKey *SecureString) (*Client, error) {
	cfg := Config{
		APIKey:       apiKey,
		DefaultModel: DefaultModel,
		Timeout:      DefaultTimeout,
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return newClientFromConn(conn, cfg), nil
}

// newClientFromConn initializes all service clients from a connection.
func newClientFromConn(conn *grpc.ClientConn, cfg Config) *Client {
	return &Client{
		conn:      conn,
		config:    cfg,
		chat:      v1.NewChatClient(conn),
		models:    v1.NewModelsClient(conn),
		embedder:  v1.NewEmbedderClient(conn),
		tokenizer: v1.NewTokenizeClient(conn),
		auth:      v1.NewAuthClient(conn),
		sampler:   v1.NewSampleClient(conn),
		image:     v1.NewImageClient(conn),
		documents: v1.NewDocumentsClient(conn),
		batch:     v1.NewBatchMgmtClient(conn),
	}
}

// Close closes the client connection and clears the API key from memory.
func (c *Client) Close() error {
	if c.config.APIKey != nil {
		c.config.APIKey.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// DefaultModel returns the default model configured for this client.
func (c *Client) DefaultModel() string {
	return c.config.DefaultModel
}

// DefaultImageModel returns the default image model.
func (c *Client) DefaultImageModel() string {
	return DefaultImageModel
}

// Timeout returns the default timeout configured for this client.
func (c *Client) Timeout() time.Duration {
	return c.config.Timeout
}

// withTimeout returns a context with the client's default timeout if the
// provided context doesn't already have a deadline.
func (c *Client) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, c.config.Timeout)
}

// bearerAuth implements grpc.PerRPCCredentials for bearer token auth.
type bearerAuth struct {
	apiKey *SecureString
}

// GetRequestMetadata returns the authorization header.
func (b *bearerAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	if b.apiKey == nil || b.apiKey.IsZero() {
		return nil, &Error{
			Code:    ErrAuth,
			Message: "API key is not set or has been closed",
		}
	}
	return map[string]string{
		"authorization": "Bearer " + b.apiKey.Value(),
	}, nil
}

// RequireTransportSecurity indicates that TLS is required.
func (b *bearerAuth) RequireTransportSecurity() bool {
	return true
}
