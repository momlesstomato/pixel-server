package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/user/domain"
)

// TestRepositoryRecordRespectLimitAndCounters verifies respect limits and counter updates.
func TestRepositoryRecordRespectLimitAndCounters(t *testing.T) {
	repository := openRepository(t)
	actor, _ := repository.Create(context.Background(), "actor")
	target, _ := repository.Create(context.Background(), "target")
	now := time.Date(2026, time.March, 13, 10, 0, 0, 0, time.UTC)
	for idx := 0; idx < 3; idx++ {
		received, err := repository.RecordRespect(context.Background(), actor.ID, target.ID, domain.RespectTargetUser, now)
		if err != nil {
			t.Fatalf("expected respect save success, got %v", err)
		}
		if received != idx+1 {
			t.Fatalf("expected respects_received=%d got %d", idx+1, received)
		}
	}
	remaining, err := repository.RemainingRespects(context.Background(), actor.ID, domain.RespectTargetUser, now)
	if err != nil {
		t.Fatalf("expected remaining query success, got %v", err)
	}
	if remaining != 0 {
		t.Fatalf("expected zero remaining respects, got %d", remaining)
	}
	if _, err := repository.RecordRespect(context.Background(), actor.ID, target.ID, domain.RespectTargetUser, now); !errors.Is(err, domain.ErrRespectLimitReached) {
		t.Fatalf("expected respect limit error, got %v", err)
	}
	nextDayRemaining, err := repository.RemainingRespects(context.Background(), actor.ID, domain.RespectTargetUser, now.Add(24*time.Hour))
	if err != nil {
		t.Fatalf("expected next day remaining query success, got %v", err)
	}
	if nextDayRemaining != domain.DefaultDailyRespects {
		t.Fatalf("expected full next-day respects, got %d", nextDayRemaining)
	}
}

// TestRepositoryRecordRespectValidation verifies actor and target existence validation.
func TestRepositoryRecordRespectValidation(t *testing.T) {
	repository := openRepository(t)
	actor, _ := repository.Create(context.Background(), "actor")
	now := time.Now().UTC()
	if _, err := repository.RecordRespect(context.Background(), 999, actor.ID, domain.RespectTargetUser, now); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected actor user not found, got %v", err)
	}
	if _, err := repository.RecordRespect(context.Background(), actor.ID, 999, domain.RespectTargetUser, now); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected target user not found, got %v", err)
	}
}
