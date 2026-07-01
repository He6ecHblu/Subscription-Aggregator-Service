package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"subscription-aggregator-service/internal/domain"
	"subscription-aggregator-service/internal/repository"

	"github.com/google/uuid"
)

func TestSubscriptionServiceCalculateTotal(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	anotherUserID := uuid.New()
	serviceName := "Yandex Plus"

	repo := &fakeSubscriptionRepository{
		listForTotal: []domain.Subscription{
			newSubscription(t, "Yandex Plus", 400, userID, "07-2025", stringPtr("12-2025")),
			newSubscription(t, "Yandex Plus", 100, userID, "10-2025", nil),
			newSubscription(t, "Netflix", 500, anotherUserID, "01-2025", stringPtr("08-2025")),
		},
	}

	svc := NewSubscriptionService(repo)

	got, err := svc.CalculateTotal(context.Background(), CalculateTotalInput{
		From:        "09-2025",
		To:          "12-2025",
		UserID:      userID.String(),
		ServiceName: serviceName,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Total != 1900 {
		t.Fatalf("got total %d, want 1900", got.Total)
	}

	if got.Currency != "RUB" {
		t.Fatalf("got currency %q, want RUB", got.Currency)
	}

	if repo.lastTotalFilter.UserID == nil || *repo.lastTotalFilter.UserID != userID {
		t.Fatalf("user filter was not passed to repository")
	}

	if repo.lastTotalFilter.ServiceName == nil || *repo.lastTotalFilter.ServiceName != serviceName {
		t.Fatalf("service name filter was not passed to repository")
	}
}

func TestSubscriptionServiceCalculateTotalRejectsInvalidPeriod(t *testing.T) {
	t.Parallel()

	svc := NewSubscriptionService(&fakeSubscriptionRepository{})

	_, err := svc.CalculateTotal(context.Background(), CalculateTotalInput{
		From: "12-2025",
		To:   "07-2025",
	})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("got error %v, want validation error", err)
	}
}

func TestSubscriptionServiceCreateValidatesInput(t *testing.T) {
	t.Parallel()

	svc := NewSubscriptionService(&fakeSubscriptionRepository{})

	_, err := svc.Create(context.Background(), CreateSubscriptionInput{
		ServiceName: " ",
		Price:       400,
		UserID:      uuid.NewString(),
		StartDate:   "07-2025",
	})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("got error %v, want validation error", err)
	}

	_, err = svc.Create(context.Background(), CreateSubscriptionInput{
		ServiceName: "Yandex Plus",
		Price:       0,
		UserID:      uuid.NewString(),
		StartDate:   "07-2025",
	})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("got error %v, want validation error", err)
	}

	endDate := "06-2025"
	_, err = svc.Create(context.Background(), CreateSubscriptionInput{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      uuid.NewString(),
		StartDate:   "07-2025",
		EndDate:     &endDate,
	})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("got error %v, want validation error", err)
	}
}

func TestSubscriptionServiceMapsRepositoryNotFound(t *testing.T) {
	t.Parallel()

	svc := NewSubscriptionService(&fakeSubscriptionRepository{getErr: repository.ErrNotFound})

	_, err := svc.GetByID(context.Background(), uuid.NewString())
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("got error %v, want not found error", err)
	}
}

type fakeSubscriptionRepository struct {
	listForTotal    []domain.Subscription
	lastTotalFilter domain.TotalFilter
	getErr          error
}

func (r *fakeSubscriptionRepository) Create(ctx context.Context, sub domain.Subscription) (domain.Subscription, error) {
	sub.ID = uuid.New()
	sub.CreatedAt = time.Now()
	sub.UpdatedAt = sub.CreatedAt
	return sub, nil
}

func (r *fakeSubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Subscription, error) {
	if r.getErr != nil {
		return domain.Subscription{}, r.getErr
	}

	return domain.Subscription{ID: id}, nil
}

func (r *fakeSubscriptionRepository) Update(ctx context.Context, sub domain.Subscription) (domain.Subscription, error) {
	return sub, nil
}

func (r *fakeSubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (r *fakeSubscriptionRepository) List(ctx context.Context, filter domain.SubscriptionFilter) ([]domain.Subscription, error) {
	return nil, nil
}

func (r *fakeSubscriptionRepository) ListForTotal(ctx context.Context, filter domain.TotalFilter) ([]domain.Subscription, error) {
	r.lastTotalFilter = filter
	return r.listForTotal, nil
}

func newSubscription(t *testing.T, serviceName string, price int, userID uuid.UUID, startDate string, endDate *string) domain.Subscription {
	t.Helper()

	start, err := domain.ParseMonth(startDate)
	if err != nil {
		t.Fatalf("parse start date: %v", err)
	}

	var end *domain.Month
	if endDate != nil {
		parsedEnd, err := domain.ParseMonth(*endDate)
		if err != nil {
			t.Fatalf("parse end date: %v", err)
		}
		end = &parsedEnd
	}

	return domain.Subscription{
		ID:          uuid.New(),
		ServiceName: serviceName,
		Price:       price,
		UserID:      userID,
		StartDate:   start,
		EndDate:     end,
	}
}

func stringPtr(value string) *string {
	return &value
}
