package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	permissionhttpapi "github.com/momlesstomato/pixel-server/pkg/permission/adapter/httpapi"
	permissionapplication "github.com/momlesstomato/pixel-server/pkg/permission/application"
	permissiondomain "github.com/momlesstomato/pixel-server/pkg/permission/domain"
)

// TestRegisterRoutesValidations verifies route registration validation behavior.
func TestRegisterRoutesValidations(t *testing.T) {
	if err := permissionhttpapi.RegisterRoutes(nil, serviceStub{}); err == nil {
		t.Fatalf("expected nil module validation failure")
	}
	module := corehttp.New(corehttp.Options{})
	if err := permissionhttpapi.RegisterRoutes(module, nil); err == nil {
		t.Fatalf("expected nil service validation failure")
	}
}

// TestGroupRoutes verifies group endpoint behavior.
func TestGroupRoutes(t *testing.T) {
	module := corehttp.New(corehttp.Options{})
	if err := permissionhttpapi.RegisterRoutes(module, serviceStub{}); err != nil {
		t.Fatalf("expected route registration success, got %v", err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/groups", nil)
	response, err := module.App().Test(request)
	if err != nil {
		t.Fatalf("expected request success, got %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}
}

// TestGroupErrorMapping verifies not-found error mapping behavior.
func TestGroupErrorMapping(t *testing.T) {
	module := corehttp.New(corehttp.Options{})
	if err := permissionhttpapi.RegisterRoutes(module, serviceStub{notFound: true}); err != nil {
		t.Fatalf("expected route registration success, got %v", err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/groups/9", nil)
	response, err := module.App().Test(request)
	if err != nil {
		t.Fatalf("expected request success, got %v", err)
	}
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", response.StatusCode)
	}
}

// TestUserGroupRoutes verifies user group assignment endpoint behavior.
func TestUserGroupRoutes(t *testing.T) {
	module := corehttp.New(corehttp.Options{})
	if err := permissionhttpapi.RegisterRoutes(module, serviceStub{}); err != nil {
		t.Fatalf("expected route registration success, got %v", err)
	}
	body := strings.NewReader(`{"group_id":1}`)
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/users/1/group", body)
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	response, err := module.App().Test(request)
	if err != nil {
		t.Fatalf("expected request success, got %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}
}

// serviceStub defines deterministic HTTP service behavior for route tests.
type serviceStub struct {
	notFound bool
}

// ListGroups returns deterministic group payload.
func (service serviceStub) ListGroups(context.Context) ([]permissionapplication.GroupDetails, error) {
	return []permissionapplication.GroupDetails{{Group: permissiondomain.Group{ID: 1, Name: "default"}}}, nil
}

// GetGroup returns deterministic group payload.
func (service serviceStub) GetGroup(context.Context, int) (permissionapplication.GroupDetails, error) {
	if service.notFound {
		return permissionapplication.GroupDetails{}, permissiondomain.ErrGroupNotFound
	}
	return permissionapplication.GroupDetails{Group: permissiondomain.Group{ID: 1, Name: "default"}}, nil
}

// CreateGroup returns deterministic group payload.
func (service serviceStub) CreateGroup(context.Context, permissionapplication.CreateGroupInput) (permissionapplication.GroupDetails, error) {
	return permissionapplication.GroupDetails{Group: permissiondomain.Group{ID: 2, Name: "new"}}, nil
}

// UpdateGroup returns deterministic group payload.
func (service serviceStub) UpdateGroup(context.Context, int, permissiondomain.GroupPatch) (permissionapplication.GroupDetails, error) {
	return permissionapplication.GroupDetails{Group: permissiondomain.Group{ID: 1, Name: "default"}}, nil
}

// DeleteGroup returns deterministic delete behavior.
func (service serviceStub) DeleteGroup(context.Context, int) error { return nil }

// AddPermissions returns deterministic group payload.
func (service serviceStub) AddPermissions(context.Context, int, []string) (permissionapplication.GroupDetails, error) {
	return permissionapplication.GroupDetails{Group: permissiondomain.Group{ID: 1, Name: "default"}, Permissions: []string{"perk.safe_chat"}}, nil
}

// RemovePermission returns deterministic group payload.
func (service serviceStub) RemovePermission(context.Context, int, string) (permissionapplication.GroupDetails, error) {
	return permissionapplication.GroupDetails{Group: permissiondomain.Group{ID: 1, Name: "default"}}, nil
}

// ReplaceUserGroups returns deterministic access payload.
func (service serviceStub) ReplaceUserGroups(context.Context, int, []int) (permissiondomain.Access, error) {
	return permissiondomain.Access{UserID: 1, PrimaryGroup: permissiondomain.Group{ID: 1, Name: "default"}}, nil
}
