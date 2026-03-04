package memory

import (
	"context"
	"sync"
	"sync/atomic"

	"pixel-server/pkg/user"
)

// RoleRepo is an in-memory implementation of user.RoleRepository.
// Suitable for unit tests and local development; no persistence across restarts.
type RoleRepo struct {
	mu          sync.RWMutex
	roles       map[int32]*user.Role
	assignments map[int32][]int32 // userID → []roleID
	nextID      atomic.Int32
}

// NewRoleRepo returns an empty, ready-to-use in-memory RoleRepo.
func NewRoleRepo() *RoleRepo {
	return &RoleRepo{
		roles:       make(map[int32]*user.Role),
		assignments: make(map[int32][]int32),
	}
}

// GetByID loads a role by its ID.
func (r *RoleRepo) GetByID(_ context.Context, id int32) (*user.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	role, ok := r.roles[id]
	if !ok {
		return nil, user.ErrRoleNotFound
	}
	return copyRole(role), nil
}

// GetAll returns all defined roles.
func (r *RoleRepo) GetAll(_ context.Context) ([]*user.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*user.Role, 0, len(r.roles))
	for _, role := range r.roles {
		out = append(out, copyRole(role))
	}
	return out, nil
}

// GetForUser returns all roles assigned to the given user.
func (r *RoleRepo) GetForUser(_ context.Context, userID int32) ([]*user.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	roleIDs := r.assignments[userID]
	out := make([]*user.Role, 0, len(roleIDs))
	for _, rid := range roleIDs {
		if role, ok := r.roles[rid]; ok {
			out = append(out, copyRole(role))
		}
	}
	return out, nil
}

// Create stores a new role, assigning an ID if zero.
func (r *RoleRepo) Create(_ context.Context, role *user.Role) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if role.ID == 0 {
		role.ID = r.nextID.Add(1)
	}
	r.roles[role.ID] = copyRole(role)
	return nil
}

// Update overwrites an existing role.
func (r *RoleRepo) Update(_ context.Context, role *user.Role) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.roles[role.ID]; !ok {
		return user.ErrRoleNotFound
	}
	r.roles[role.ID] = copyRole(role)
	return nil
}

// Delete removes a role and revokes it from all users.
func (r *RoleRepo) Delete(_ context.Context, id int32) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.roles[id]; !ok {
		return user.ErrRoleNotFound
	}
	delete(r.roles, id)
	for uid, ids := range r.assignments {
		filtered := ids[:0]
		for _, rid := range ids {
			if rid != id {
				filtered = append(filtered, rid)
			}
		}
		r.assignments[uid] = filtered
	}
	return nil
}

// AssignRole grants a role to a user. Idempotent.
func (r *RoleRepo) AssignRole(_ context.Context, userID, roleID int32) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.roles[roleID]; !ok {
		return user.ErrRoleNotFound
	}
	for _, existing := range r.assignments[userID] {
		if existing == roleID {
			return nil
		}
	}
	r.assignments[userID] = append(r.assignments[userID], roleID)
	return nil
}

// RevokeRole removes a role from a user. Idempotent.
func (r *RoleRepo) RevokeRole(_ context.Context, userID, roleID int32) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	ids := r.assignments[userID]
	filtered := ids[:0]
	for _, id := range ids {
		if id != roleID {
			filtered = append(filtered, id)
		}
	}
	r.assignments[userID] = filtered
	return nil
}

// HasRole reports whether a role is currently assigned to a user.
func (r *RoleRepo) HasRole(_ context.Context, userID, roleID int32) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, id := range r.assignments[userID] {
		if id == roleID {
			return true, nil
		}
	}
	return false, nil
}

// copyRole returns a deep copy of a role to prevent aliasing.
func copyRole(role *user.Role) *user.Role {
	cp := *role
	cp.Permissions = append([]string(nil), role.Permissions...)
	return &cp
}

// compile-time interface check.
var _ user.RoleRepository = (*RoleRepo)(nil)
