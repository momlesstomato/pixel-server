package memory

import (
	"context"
	"sync"
	"sync/atomic"

	"pixel-server/pkg/user"
)

// UserRepo is an in-memory fake that implements user.Repository.
// It is intended for unit tests only.
type UserRepo struct {
	mu       sync.RWMutex
	byID     map[int32]*user.User
	byName   map[string]*user.User
	settings map[int32]*user.Settings
	nextID   atomic.Int32
}

// NewUserRepo returns a ready-to-use in-memory UserRepo.
func NewUserRepo() *UserRepo {
	return &UserRepo{
		byID:     make(map[int32]*user.User),
		byName:   make(map[string]*user.User),
		settings: make(map[int32]*user.Settings),
	}
}

func (r *UserRepo) GetByID(_ context.Context, id int32) (*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.byID[id]
	if !ok {
		return nil, user.ErrNotFound
	}
	cp := *u
	return &cp, nil
}

func (r *UserRepo) GetByUsername(_ context.Context, username string) (*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.byName[username]
	if !ok {
		return nil, user.ErrNotFound
	}
	cp := *u
	return &cp, nil
}

func (r *UserRepo) Create(_ context.Context, u *user.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.byName[u.Username]; exists {
		return user.ErrAlreadyExists
	}
	if u.ID == 0 {
		u.ID = r.nextID.Add(1)
	}
	cp := *u
	r.byID[cp.ID] = &cp
	r.byName[cp.Username] = &cp
	return nil
}

func (r *UserRepo) Update(_ context.Context, u *user.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	old, ok := r.byID[u.ID]
	if !ok {
		return user.ErrNotFound
	}
	delete(r.byName, old.Username)
	cp := *u
	r.byID[cp.ID] = &cp
	r.byName[cp.Username] = &cp
	return nil
}

func (r *UserRepo) GetSettings(_ context.Context, userID int32) (*user.Settings, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.settings[userID]
	if !ok {
		return nil, user.ErrNotFound
	}
	cp := *s
	return &cp, nil
}

func (r *UserRepo) UpdateSettings(_ context.Context, s *user.Settings) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byID[s.UserID]; !ok {
		return user.ErrNotFound
	}
	cp := *s
	r.settings[cp.UserID] = &cp
	return nil
}

func (r *UserRepo) SetOnline(_ context.Context, userID int32, online bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.byID[userID]
	if !ok {
		return user.ErrNotFound
	}
	u.Online = online
	return nil
}

// compile-time interface check.
var _ user.Repository = (*UserRepo)(nil)
