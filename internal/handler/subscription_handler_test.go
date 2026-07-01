package handler

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"subscription-aggregator-service/internal/domain"
	"subscription-aggregator-service/internal/service"

	"github.com/google/uuid"
)

func TestSubscriptionHandlerCreate(t *testing.T) {
	t.Parallel()

	fake := &fakeSubscriptionService{}
	h := NewSubscriptionHandler(fake, slog.New(slog.NewTextHandler(io.Discard, nil)))

	body := `{
		"service_name": "Yandex Plus",
		"price": 400,
		"user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		"start_date": "07-2025",
		"end_date": "12-2025"
	}`

	req := httptest.NewRequest(http.MethodPost, "/subscriptions", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("got status %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	if fake.createInput.ServiceName != "Yandex Plus" {
		t.Fatalf("service input was not populated")
	}

	if !strings.Contains(rec.Body.String(), `"service_name":"Yandex Plus"`) {
		t.Fatalf("response does not contain subscription: %s", rec.Body.String())
	}
}

func TestSubscriptionHandlerValidationError(t *testing.T) {
	t.Parallel()

	fake := &fakeSubscriptionService{
		totalErr: &service.Error{
			Kind:    service.ErrValidation,
			Message: "from must have MM-YYYY format",
		},
	}
	h := NewSubscriptionHandler(fake, slog.New(slog.NewTextHandler(io.Discard, nil)))

	req := httptest.NewRequest(http.MethodGet, "/subscriptions/total?from=bad&to=12-2025", nil)
	rec := httptest.NewRecorder()

	h.CalculateTotal(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", rec.Code, http.StatusBadRequest)
	}

	if !strings.Contains(rec.Body.String(), `"error":"validation_error"`) {
		t.Fatalf("response does not contain validation error: %s", rec.Body.String())
	}
}

func TestSubscriptionHandlerNotFoundError(t *testing.T) {
	t.Parallel()

	fake := &fakeSubscriptionService{
		getErr: &service.Error{
			Kind:    service.ErrNotFound,
			Message: "subscription not found",
		},
	}
	h := NewSubscriptionHandler(fake, slog.New(slog.NewTextHandler(io.Discard, nil)))

	req := httptest.NewRequest(http.MethodGet, "/subscriptions/"+uuid.NewString(), nil)
	rec := httptest.NewRecorder()

	h.GetByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d", rec.Code, http.StatusNotFound)
	}

	if !strings.Contains(rec.Body.String(), `"error":"not_found"`) {
		t.Fatalf("response does not contain not_found error: %s", rec.Body.String())
	}
}

type fakeSubscriptionService struct {
	createInput service.CreateSubscriptionInput
	totalErr    error
	getErr      error
}

func (s *fakeSubscriptionService) Create(ctx context.Context, input service.CreateSubscriptionInput) (domain.Subscription, error) {
	s.createInput = input
	return testSubscription(), nil
}

func (s *fakeSubscriptionService) GetByID(ctx context.Context, id string) (domain.Subscription, error) {
	if s.getErr != nil {
		return domain.Subscription{}, s.getErr
	}

	return testSubscription(), nil
}

func (s *fakeSubscriptionService) Update(ctx context.Context, input service.UpdateSubscriptionInput) (domain.Subscription, error) {
	return testSubscription(), nil
}

func (s *fakeSubscriptionService) Delete(ctx context.Context, id string) error {
	return nil
}

func (s *fakeSubscriptionService) List(ctx context.Context, input service.ListSubscriptionsInput) ([]domain.Subscription, error) {
	return []domain.Subscription{testSubscription()}, nil
}

func (s *fakeSubscriptionService) CalculateTotal(ctx context.Context, input service.CalculateTotalInput) (service.CalculateTotalResult, error) {
	if s.totalErr != nil {
		return service.CalculateTotalResult{}, s.totalErr
	}

	from, err := domain.ParseMonth("07-2025")
	if err != nil {
		return service.CalculateTotalResult{}, err
	}

	to, err := domain.ParseMonth("12-2025")
	if err != nil {
		return service.CalculateTotalResult{}, err
	}

	return service.CalculateTotalResult{
		Total:    2400,
		Currency: "RUB",
		From:     from,
		To:       to,
	}, nil
}

func testSubscription() domain.Subscription {
	start, err := domain.ParseMonth("07-2025")
	if err != nil {
		panic(err)
	}

	end, err := domain.ParseMonth("12-2025")
	if err != nil {
		panic(err)
	}

	return domain.Subscription{
		ID:          uuid.MustParse("2b0a9d62-43ec-4ef5-8f7a-49ec0bdf7611"),
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba"),
		StartDate:   start,
		EndDate:     &end,
		CreatedAt:   time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC),
	}
}

var _ SubscriptionService = (*fakeSubscriptionService)(nil)
