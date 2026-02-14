package xai

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"
)

// APIKeyStatus represents the status of an API key.
type APIKeyStatus int

const (
	// APIKeyActive indicates the key is active and can make requests.
	APIKeyActive APIKeyStatus = iota
	// APIKeyDisabled indicates the key has been disabled.
	APIKeyDisabled
	// APIKeyBlocked indicates the key has been blocked.
	APIKeyBlocked
	// APIKeyTeamBlocked indicates the key's team has been blocked.
	APIKeyTeamBlocked
)

// String returns a human-readable status string.
func (s APIKeyStatus) String() string {
	switch s {
	case APIKeyActive:
		return "active"
	case APIKeyDisabled:
		return "disabled"
	case APIKeyBlocked:
		return "blocked"
	case APIKeyTeamBlocked:
		return "team_blocked"
	default:
		return "unknown"
	}
}

// APIKeyInfo contains information about an API key.
type APIKeyInfo struct {
	// RedactedKey is a partially redacted version of the API key.
	RedactedKey string
	// KeyID is the unique identifier for this API key.
	KeyID string
	// Name is the human-readable name for the API key.
	Name string
	// UserID is the ID of the user who created this key.
	UserID string
	// TeamID is the ID of the team this key belongs to.
	TeamID string
	// ACLs are the access control lists (permissions) for this key.
	ACLs []string
	// Status is the current status of the key.
	Status APIKeyStatus
	// CreatedAt is when the key was created.
	CreatedAt time.Time
	// ModifiedAt is when the key was last modified.
	ModifiedAt time.Time
	// ModifiedBy is the ID of the user who last modified the key.
	ModifiedBy string
}

// IsActive returns true if the API key is active and can make requests.
func (k *APIKeyInfo) IsActive() bool {
	return k.Status == APIKeyActive
}

// HasACL checks if the key has a specific ACL permission.
func (k *APIKeyInfo) HasACL(acl string) bool {
	for _, a := range k.ACLs {
		if a == acl {
			return true
		}
	}
	return false
}

// GetAPIKeyInfo retrieves information about the current API key.
func (c *Client) GetAPIKeyInfo(ctx context.Context) (*APIKeyInfo, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	resp, err := c.auth.GetApiKeyInfo(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, FromGRPCError(err)
	}

	info := &APIKeyInfo{
		RedactedKey: resp.GetRedactedApiKey(),
		KeyID:       resp.GetApiKeyId(),
		Name:        resp.GetName(),
		UserID:      resp.GetUserId(),
		TeamID:      resp.GetTeamId(),
		ACLs:        resp.GetAcls(),
		ModifiedBy:  resp.GetModifiedBy(),
	}

	// Determine status
	switch {
	case resp.GetDisabled():
		info.Status = APIKeyDisabled
	case resp.GetApiKeyBlocked():
		info.Status = APIKeyBlocked
	case resp.GetTeamBlocked():
		info.Status = APIKeyTeamBlocked
	default:
		info.Status = APIKeyActive
	}

	// Convert timestamps
	if resp.GetCreateTime() != nil {
		info.CreatedAt = resp.GetCreateTime().AsTime()
	}
	if resp.GetModifyTime() != nil {
		info.ModifiedAt = resp.GetModifyTime().AsTime()
	}

	return info, nil
}
