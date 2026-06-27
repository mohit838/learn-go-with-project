package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mohit838/learn-go-with-project/internal/notification/domain"
)

func TestSubmitRunsThroughWorkerPool(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	service := NewService(2, 4)
	service.Start(ctx)
	defer service.Stop()

	notification, err := service.Submit(ctx, domain.UserContext{
		UserID:   "user_1",
		TenantID: "tenant_1",
		Role:     "admin",
	}, SubmitRequest{
		Recipient: "dev@example.com",
		Message:   "learn channels",
	})
	if err != nil {
		t.Fatalf("submit notification: %v", err)
	}

	deadline := time.After(2 * time.Second)
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			t.Fatal("notification was not delivered before timeout")
		case <-ticker.C:
			got, err := service.Find(notification.ID)
			if err != nil {
				t.Fatalf("find notification: %v", err)
			}
			if got.Status == domain.StatusDelivered {
				if got.WorkerID == 0 {
					t.Fatal("delivered notification should record the worker id")
				}
				return
			}
		}
	}
}

func TestSubmitValidatesInput(t *testing.T) {
	service := NewService(1, 1)

	_, err := service.Submit(context.Background(), domain.UserContext{}, SubmitRequest{})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}

	_, err = service.Submit(context.Background(), domain.UserContext{}, SubmitRequest{
		Recipient: "dev@example.com",
		Message:   "hello",
	})
	if !errors.Is(err, ErrMissingUser) {
		t.Fatalf("expected ErrMissingUser, got %v", err)
	}
}
