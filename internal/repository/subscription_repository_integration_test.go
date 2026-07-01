package repository

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"subscription-aggregator-service/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestSubscriptionRepositoryIntegration(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	defer pool.Close()

	repo := NewSubscriptionRepository(pool)
	userID := uuid.New()
	start := mustParseMonth(t, "07-2025")
	end := mustParseMonth(t, "12-2025")

	created, err := repo.Create(ctx, domain.Subscription{
		ServiceName: "Integration Test Service",
		Price:       400,
		UserID:      userID,
		StartDate:   start,
		EndDate:     &end,
	})
	if err != nil {
		t.Fatalf("create subscription: %v", err)
	}
	defer func() {
		_ = repo.Delete(context.Background(), created.ID)
	}()

	if created.ID == uuid.Nil {
		t.Fatalf("created subscription has empty ID")
	}

	got, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get subscription: %v", err)
	}
	if got.ServiceName != created.ServiceName {
		t.Fatalf("got service name %q, want %q", got.ServiceName, created.ServiceName)
	}

	serviceName := "Integration Test Service"
	list, err := repo.List(ctx, domain.SubscriptionFilter{
		UserID:      &userID,
		ServiceName: &serviceName,
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("list subscriptions: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("got %d listed subscriptions, want 1", len(list))
	}

	forTotal, err := repo.ListForTotal(ctx, domain.TotalFilter{
		From:        mustParseMonth(t, "08-2025"),
		To:          mustParseMonth(t, "09-2025"),
		UserID:      &userID,
		ServiceName: &serviceName,
	})
	if err != nil {
		t.Fatalf("list subscriptions for total: %v", err)
	}
	if len(forTotal) != 1 {
		t.Fatalf("got %d subscriptions for total, want 1", len(forTotal))
	}

	created.Price = 500
	updated, err := repo.Update(ctx, created)
	if err != nil {
		t.Fatalf("update subscription: %v", err)
	}
	if updated.Price != 500 {
		t.Fatalf("got updated price %d, want 500", updated.Price)
	}

	if err := repo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("delete subscription: %v", err)
	}

	_, err = repo.GetByID(ctx, created.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("got error %v, want ErrNotFound", err)
	}
}

func mustParseMonth(t *testing.T, value string) domain.Month {
	t.Helper()

	month, err := domain.ParseMonth(value)
	if err != nil {
		t.Fatalf("parse month %q: %v", value, err)
	}

	return month
}
