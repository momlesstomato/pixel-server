package memory

import (
	"context"
	"sync"

	"pixel-server/pkg/user"
)

// IgnoreRepo is an in-memory fake that implements user.IgnoreRepository.
type IgnoreRepo struct {
	mu      sync.RWMutex
	ignores map[int32]map[int32]*user.IgnoredUser
}

// NewIgnoreRepo returns a ready-to-use in-memory IgnoreRepo.
func NewIgnoreRepo() *IgnoreRepo {
	return &IgnoreRepo{ignores: make(map[int32]map[int32]*user.IgnoredUser)}
}

func (r *IgnoreRepo) GetByUser(_ context.Context, userID int32) ([]*user.IgnoredUser, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*user.IgnoredUser
	for _, ig := range r.ignores[userID] {
		cp := *ig
		out = append(out, &cp)
	}
	return out, nil
}

func (r *IgnoreRepo) Add(_ context.Context, userID int32, ignoredUserID int32) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.ignores[userID]
	if !ok {
		m = make(map[int32]*user.IgnoredUser)
		r.ignores[userID] = m
	}
	m[ignoredUserID] = &user.IgnoredUser{
		UserID:        userID,
		IgnoredUserID: ignoredUserID,
	}
	return nil
}

func (r *IgnoreRepo) Remove(_ context.Context, userID int32, ignoredUserID int32) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	m := r.ignores[userID]
	if m != nil {
		delete(m, ignoredUserID)
	}
	return nil
}

func (r *IgnoreRepo) IsIgnored(_ context.Context, userID int32, targetID int32) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m := r.ignores[userID]
	if m == nil {
		return false, nil
	}
	_, ok := m[targetID]
	return ok, nil
}

// compile-time interface check.
var _ user.IgnoreRepository = (*IgnoreRepo)(nil)
