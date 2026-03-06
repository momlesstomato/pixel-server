package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"pixelsv/internal/auth/domain"
)

type stubService struct {
	createTicket string
	createTTL    int32
	createErr    error
	revokeErr    error
}

func (s *stubService) CreateTicket(userID int32, ttlSeconds int32) (string, int32, error) {
	if s.createErr != nil {
		return "", 0, s.createErr
	}
	return s.createTicket, s.createTTL, nil
}

func (s *stubService) RevokeTicket(ticket string) error {
	return s.revokeErr
}

// TestRegisterRoutes validates auth ticket route behavior.
func TestRegisterRoutes(t *testing.T) {
	app := fiber.New()
	service := &stubService{createTicket: "abc", createTTL: 300}
	RegisterRoutes(app, service, "secret")
	body, _ := json.Marshal(CreateTicketRequest{UserID: 1})
	request := httptest.NewRequest("POST", "/api/v1/auth/tickets", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response, _ := app.Test(request)
	if response.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected unauthorized status, got %d", response.StatusCode)
	}
	request = httptest.NewRequest("POST", "/api/v1/auth/tickets", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-Key", "secret")
	response, _ = app.Test(request)
	if response.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected created status, got %d", response.StatusCode)
	}
	deleteRequest := httptest.NewRequest("DELETE", "/api/v1/auth/tickets/t1", nil)
	deleteRequest.Header.Set("X-API-Key", "secret")
	deleteResponse, _ := app.Test(deleteRequest)
	if deleteResponse.StatusCode != fiber.StatusNoContent {
		t.Fatalf("expected no content status, got %d", deleteResponse.StatusCode)
	}
}

// TestRegisterRoutesErrors validates invalid request and service error mapping.
func TestRegisterRoutesErrors(t *testing.T) {
	app := fiber.New()
	service := &stubService{createErr: domain.ErrInvalidUserID, revokeErr: domain.ErrInvalidTicket}
	RegisterRoutes(app, service, "secret")
	request := httptest.NewRequest("POST", "/api/v1/auth/tickets", bytes.NewReader([]byte(`{`)))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-Key", "secret")
	response, _ := app.Test(request)
	if response.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected bad request status, got %d", response.StatusCode)
	}
	request = httptest.NewRequest("POST", "/api/v1/auth/tickets", bytes.NewReader([]byte(`{"user_id":1}`)))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-Key", "secret")
	service.createErr = errors.New("boom")
	response, _ = app.Test(request)
	if response.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("expected internal status, got %d", response.StatusCode)
	}
	deleteRequest := httptest.NewRequest("DELETE", "/api/v1/auth/tickets/t1", nil)
	deleteRequest.Header.Set("X-API-Key", "secret")
	service.revokeErr = domain.ErrInvalidTicket
	deleteResponse, _ := app.Test(deleteRequest)
	if deleteResponse.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected bad request status, got %d", deleteResponse.StatusCode)
	}
}
