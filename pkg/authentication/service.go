package authentication

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"
)

// IssueRequest defines one ticket issuance input.
type IssueRequest struct {
	// UserID defines the user identifier bound to the issued ticket.
	UserID int
	// TTL defines requested lifetime; zero uses configured default.
	TTL time.Duration
}

// IssueResult defines one issued ticket payload.
type IssueResult struct {
	// Ticket stores generated single-use authentication token.
	Ticket string
	// ExpiresAt stores token expiration timestamp.
	ExpiresAt time.Time
	// TTL stores the accepted lifetime used for storage.
	TTL time.Duration
}

// Service implements SSO ticket issue and validation workflows.
type Service struct {
	// store persists and validates issued tickets.
	store Store
	// config stores runtime ticket policy parameters.
	config Config
	// now returns current time for expiration calculation.
	now func() time.Time
	// random provides bytes for ticket generation.
	random io.Reader
}

// NewService creates an SSO service instance.
func NewService(store Store, config Config) *Service {
	return &Service{store: store, config: config, now: time.Now, random: rand.Reader}
}

// Issue generates and stores one single-use ticket.
func (service *Service) Issue(ctx context.Context, request IssueRequest) (IssueResult, error) {
	if request.UserID <= 0 {
		return IssueResult{}, fmt.Errorf("user id must be positive")
	}
	ttl, err := service.resolveTTL(request.TTL)
	if err != nil {
		return IssueResult{}, err
	}
	token := make([]byte, 16)
	if _, err := io.ReadFull(service.random, token); err != nil {
		return IssueResult{}, fmt.Errorf("generate ticket bytes: %w", err)
	}
	ticket := hex.EncodeToString(token)
	if err := service.store.Store(ctx, ticket, request.UserID, ttl); err != nil {
		return IssueResult{}, err
	}
	return IssueResult{Ticket: ticket, ExpiresAt: service.now().Add(ttl), TTL: ttl}, nil
}

// Validate consumes one ticket and returns its user ID.
func (service *Service) Validate(ctx context.Context, ticket string) (int, error) {
	trimmed := strings.TrimSpace(ticket)
	if trimmed == "" {
		return 0, fmt.Errorf("ticket is required")
	}
	return service.store.Validate(ctx, trimmed)
}

// resolveTTL normalizes and validates requested TTL against policy limits.
func (service *Service) resolveTTL(requested time.Duration) (time.Duration, error) {
	defaultTTL := time.Duration(service.config.DefaultTTLSeconds) * time.Second
	if defaultTTL <= 0 {
		defaultTTL = 5 * time.Minute
	}
	maxTTL := time.Duration(service.config.MaxTTLSeconds) * time.Second
	if maxTTL <= 0 {
		maxTTL = 30 * time.Minute
	}
	ttl := requested
	if ttl <= 0 {
		ttl = defaultTTL
	}
	if ttl > maxTTL {
		return 0, fmt.Errorf("ttl exceeds maximum of %s", maxTTL)
	}
	return ttl, nil
}
