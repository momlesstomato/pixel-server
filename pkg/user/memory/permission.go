package memory

import (
	"context"
	"sync"

	"pixel-server/pkg/user"
)

// PermissionRepo is an in-memory implementation of user.PermissionRepository.
// Suitable for unit tests and local development; no persistence across restarts.
type PermissionRepo struct {
	mu    sync.RWMutex
	perms map[string]*user.Permission
}

// NewPermissionRepo returns an empty, ready-to-use in-memory PermissionRepo.
func NewPermissionRepo() *PermissionRepo {
	return &PermissionRepo{perms: make(map[string]*user.Permission)}
}

// GetAll returns all registered permissions.
func (p *PermissionRepo) GetAll(_ context.Context) ([]*user.Permission, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]*user.Permission, 0, len(p.perms))
	for _, perm := range p.perms {
		cp := *perm
		out = append(out, &cp)
	}
	return out, nil
}

// GetByName loads a permission by name.
func (p *PermissionRepo) GetByName(_ context.Context, name string) (*user.Permission, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	perm, ok := p.perms[name]
	if !ok {
		return nil, user.ErrPermissionNotFound
	}
	cp := *perm
	return &cp, nil
}

// Create registers or overwrites a permission by name.
func (p *PermissionRepo) Create(_ context.Context, perm *user.Permission) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	cp := *perm
	p.perms[cp.Name] = &cp
	return nil
}

// Delete removes a permission definition by name.
func (p *PermissionRepo) Delete(_ context.Context, name string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.perms[name]; !ok {
		return user.ErrPermissionNotFound
	}
	delete(p.perms, name)
	return nil
}

// compile-time interface check.
var _ user.PermissionRepository = (*PermissionRepo)(nil)
