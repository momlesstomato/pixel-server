package application

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/momlesstomato/pixel-server/core/broadcast"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	sdk "github.com/momlesstomato/pixel-sdk"
	"github.com/momlesstomato/pixel-server/pkg/messenger/domain"
	sessionnotification "github.com/momlesstomato/pixel-server/pkg/session/application/notification"
)

// Config defines messenger service runtime configuration.
type Config struct {
	// MaxFriends stores the normal friend list capacity.
	MaxFriends int `mapstructure:"max_friends" default:"200"`
	// MaxFriendsVIP stores the VIP friend list capacity.
	MaxFriendsVIP int `mapstructure:"max_friends_vip" default:"500"`
	// FloodCooldownMs stores the minimum milliseconds between messages.
	FloodCooldownMs int `mapstructure:"flood_cooldown_ms" default:"750"`
	// FloodViolations stores the violation count threshold that triggers a mute.
	FloodViolations int `mapstructure:"flood_violations" default:"4"`
	// FloodMuteSeconds stores the mute duration in seconds after flood threshold.
	FloodMuteSeconds int `mapstructure:"flood_mute_seconds" default:"20"`
	// OfflineMsgTTLDays stores the offline message retention period in days.
	OfflineMsgTTLDays int `mapstructure:"offline_msg_ttl_days" default:"30"`
	// MessageLogTTLDays stores the message log retention period in days.
	MessageLogTTLDays int `mapstructure:"message_log_ttl_days" default:"30"`
	// PurgeIntervalSeconds stores how often the purge job runs in seconds.
	PurgeIntervalSeconds int `mapstructure:"purge_interval_seconds" default:"3600"`
	// FragmentSize stores the number of friends per list fragment packet.
	FragmentSize int `mapstructure:"fragment_size" default:"750"`
}

// floodState tracks message flood control state per connection.
type floodState struct {
	// lastMessage stores the last message timestamp.
	lastMessage time.Time
	// violations stores accumulated speed violations.
	violations int
	// mutedUntil stores the mute expiry timestamp; zero when not muted.
	mutedUntil time.Time
}

// Service defines messenger application use-cases.
type Service struct {
	// repository stores messenger persistence contract.
	repository domain.Repository
	// sessions stores active connection registry.
	sessions coreconnection.SessionRegistry
	// broadcaster stores cross-instance publish behavior.
	broadcaster broadcast.Broadcaster
	// fire stores optional plugin event dispatch behavior.
	fire func(sdk.Event)
	// checker stores optional permission resolution behavior.
	checker domain.PermissionChecker
	// flood guards per-connection message rate state.
	flood map[string]*floodState
	// floodMu guards the flood map.
	floodMu sync.Mutex
	// config stores runtime service configuration.
	config Config
}

// NewService creates one messenger service.
func NewService(repository domain.Repository, sessions coreconnection.SessionRegistry, broadcaster broadcast.Broadcaster, config Config) (*Service, error) {
	if repository == nil {
		return nil, fmt.Errorf("messenger repository is required")
	}
	if sessions == nil {
		return nil, fmt.Errorf("session registry is required")
	}
	if broadcaster == nil {
		return nil, fmt.Errorf("broadcaster is required")
	}
	if config.MaxFriends <= 0 {
		config.MaxFriends = 200
	}
	if config.MaxFriendsVIP <= 0 {
		config.MaxFriendsVIP = 500
	}
	if config.FloodCooldownMs <= 0 {
		config.FloodCooldownMs = 750
	}
	if config.FloodViolations <= 0 {
		config.FloodViolations = 4
	}
	if config.FloodMuteSeconds <= 0 {
		config.FloodMuteSeconds = 20
	}
	if config.OfflineMsgTTLDays <= 0 {
		config.OfflineMsgTTLDays = 30
	}
	if config.MessageLogTTLDays <= 0 {
		config.MessageLogTTLDays = 30
	}
	if config.PurgeIntervalSeconds <= 0 {
		config.PurgeIntervalSeconds = 3600
	}
	if config.FragmentSize <= 0 {
		config.FragmentSize = 750
	}
	return &Service{
		repository: repository, sessions: sessions, broadcaster: broadcaster,
		config: config, flood: make(map[string]*floodState),
	}, nil
}

// SetEventFirer configures optional plugin event dispatch behavior.
func (service *Service) SetEventFirer(fire func(sdk.Event)) {
	service.fire = fire
}

// SetPermissionChecker configures optional permission resolution behavior.
func (service *Service) SetPermissionChecker(checker domain.PermissionChecker) {
	service.checker = checker
}

// ResolvedFriendLimit returns the effective friend limit for one user.
// Returns 0 when the user holds the unlimited permission.
func (service *Service) ResolvedFriendLimit(ctx context.Context, userID int) int {
	if service.checker != nil {
		if ok, _ := service.checker.HasPermission(ctx, userID, domain.PermFriendsUnlimited); ok {
			return 0
		}
		if ok, _ := service.checker.HasPermission(ctx, userID, domain.PermFriendsExtended); ok {
			return service.config.MaxFriendsVIP
		}
	}
	return service.config.MaxFriends
}

// userChannel returns the per-user broadcast channel key.
func userChannel(userID int) string {
	return sessionnotification.UserChannel(userID)
}

// Config returns the service runtime configuration.
func (service *Service) Config() Config { return service.config }
