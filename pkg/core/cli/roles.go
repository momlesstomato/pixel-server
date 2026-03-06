package cli

import (
	"sort"

	"pixelsv/pkg/config"
)

// roleSet stores normalized active runtime roles.
type roleSet map[string]struct{}

// newRoleSet parses a comma-separated role expression into a roleSet.
func newRoleSet(value string) (roleSet, error) {
	roles, err := config.ParseRoles(value)
	if err != nil {
		return nil, err
	}
	set := make(roleSet, len(roles))
	for _, role := range roles {
		set[role] = struct{}{}
	}
	return set, nil
}

// has reports whether a role is active.
func (r roleSet) has(role string) bool {
	_, ok := r[role]
	return ok
}

// needsHTTP reports whether this role set must expose HTTP endpoints.
func (r roleSet) needsHTTP() bool {
	return r.has("all") || r.has("gateway") || r.has("api") || r.has("jobs")
}

// needsPostgres reports whether this role set needs PostgreSQL.
func (r roleSet) needsPostgres() bool {
	if r.has("all") {
		return true
	}
	return r.has("game") || r.has("auth") || r.has("social") || r.has("navigator") || r.has("catalog") || r.has("moderation") || r.has("api") || r.has("jobs")
}

// needsRedis reports whether this role set needs Redis.
func (r roleSet) needsRedis() bool {
	if r.has("all") {
		return true
	}
	return r.has("gateway") || r.has("game") || r.has("auth") || r.has("social") || r.has("navigator") || r.has("moderation") || r.has("api") || r.has("jobs")
}

// forceLocalTransport reports whether runtime transport must stay in-process.
func (r roleSet) forceLocalTransport() bool {
	return r.has("all")
}

// names returns a sorted list of active role names.
func (r roleSet) names() []string {
	values := make([]string, 0, len(r))
	for role := range r {
		values = append(values, role)
	}
	sort.Strings(values)
	return values
}
