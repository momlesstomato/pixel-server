package tests

import (
	"context"
	"testing"

	"github.com/momlesstomato/pixel-server/pkg/messenger/domain"
)

// TestSendRequest_TargetNotFound verifies error when username does not exist.
func TestSendRequest_TargetNotFound(t *testing.T) {
	repo := repositoryStub{usernameFound: false}
	service := newTestService(repo, &sessionRegistryStub{}, &broadcasterStub{})
	_, _, err := service.SendRequest(context.Background(), "conn1", 1, "ghost")
	if err != domain.ErrTargetNotFound {
		t.Fatalf("expected ErrTargetNotFound, got %v", err)
	}
}

// TestSendRequest_SelfRequest verifies error when sending to own username.
func TestSendRequest_SelfRequest(t *testing.T) {
	repo := repositoryStub{usernameFound: true, userID: 1}
	service := newTestService(repo, &sessionRegistryStub{}, &broadcasterStub{})
	_, _, err := service.SendRequest(context.Background(), "conn1", 1, "self")
	if err != domain.ErrSelfRequest {
		t.Fatalf("expected ErrSelfRequest, got %v", err)
	}
}

// TestSendRequest_AlreadyFriends verifies error when users are already friends.
func TestSendRequest_AlreadyFriends(t *testing.T) {
	repo := repositoryStub{usernameFound: true, userID: 2, areFriends: true}
	service := newTestService(repo, &sessionRegistryStub{}, &broadcasterStub{})
	_, _, err := service.SendRequest(context.Background(), "conn1", 1, "user2")
	if err != domain.ErrAlreadyFriends {
		t.Fatalf("expected ErrAlreadyFriends, got %v", err)
	}
}

// TestSendRequest_CrossRequestAutoAccept verifies auto-accept when reverse request exists.
func TestSendRequest_CrossRequestAutoAccept(t *testing.T) {
	repo := repositoryStub{usernameFound: true, userID: 2, requestFound: true,
		request: domain.FriendRequest{ID: 10, FromUserID: 2, ToUserID: 1}}
	service := newTestService(repo, &sessionRegistryStub{}, &broadcasterStub{})
	req, accepted, err := service.SendRequest(context.Background(), "conn1", 1, "user2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !accepted {
		t.Fatal("expected auto-accepted=true")
	}
	if req.FromUserID != 1 || req.ToUserID != 2 {
		t.Fatalf("unexpected request: %+v", req)
	}
}

// TestSendRequest_Positive verifies standard request creation.
func TestSendRequest_Positive(t *testing.T) {
	repo := repositoryStub{usernameFound: true, userID: 2}
	service := newTestService(repo, &sessionRegistryStub{}, &broadcasterStub{})
	req, accepted, err := service.SendRequest(context.Background(), "conn1", 1, "user2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if accepted {
		t.Fatal("expected accepted=false for new request")
	}
	if req.FromUserID != 1 {
		t.Fatalf("unexpected fromUserID=%d", req.FromUserID)
	}
}

// TestAcceptRequest_AlreadyFriends verifies that accepting returns nil when friendship already exists.
func TestAcceptRequest_AlreadyFriends(t *testing.T) {
	repo := repositoryStub{requestFound: false, areFriends: true}
	service := newTestService(repo, &sessionRegistryStub{}, &broadcasterStub{})
	err := service.AcceptRequest(context.Background(), 99, 5)
	if err != nil {
		t.Fatalf("expected no error when already friends, got %v", err)
	}
}

// TestAcceptRequest_LateAccept verifies friendship creation when the request was removed by a prior decline.
func TestAcceptRequest_LateAccept(t *testing.T) {
	repo := repositoryStub{requestFound: false, areFriends: false}
	service := newTestService(repo, &sessionRegistryStub{}, &broadcasterStub{})
	err := service.AcceptRequest(context.Background(), 3, 5)
	if err != nil {
		t.Fatalf("expected friendship to be created for late accept, got %v", err)
	}
}

// TestAcceptRequest_Positive verifies friendship creation from request.
func TestAcceptRequest_Positive(t *testing.T) {
	req := domain.FriendRequest{ID: 1, FromUserID: 5, ToUserID: 3}
	repo := repositoryStub{request: req, requestFound: true}
	service := newTestService(repo, &sessionRegistryStub{}, &broadcasterStub{})
	err := service.AcceptRequest(context.Background(), 3, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestDeclineRequest_Positive verifies request deletion for the owning user.
func TestDeclineRequest_Positive(t *testing.T) {
	repo := repositoryStub{request: domain.FriendRequest{ID: 1, FromUserID: 5, ToUserID: 1}, requestFound: true}
	service := newTestService(repo, &sessionRegistryStub{}, &broadcasterStub{})
	err := service.DeclineRequest(context.Background(), 1, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestDeclineAllRequests_InvalidUser verifies error for zero user ID.
func TestDeclineAllRequests_InvalidUser(t *testing.T) {
	service := newTestService(repositoryStub{}, &sessionRegistryStub{}, &broadcasterStub{})
	err := service.DeclineAllRequests(context.Background(), 0)
	if err == nil {
		t.Fatal("expected error for zero userID")
	}
}
