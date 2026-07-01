package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"subscription-aggregator-service/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("subscription not found")

type SubscriptionRepository struct {
	pool *pgxpool.Pool
}

func NewSubscriptionRepository(pool *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{pool: pool}
}

func (r *SubscriptionRepository) Create(ctx context.Context, sub domain.Subscription) (domain.Subscription, error) {
	const query = `
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate.Time(),
		monthPtrTime(sub.EndDate),
	)

	created, err := scanSubscription(row)
	if err != nil {
		return domain.Subscription{}, fmt.Errorf("create subscription: %w", err)
	}

	return created, nil
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Subscription, error) {
	const query = `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE id = $1
	`

	sub, err := scanSubscription(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		return domain.Subscription{}, mapNotFound(err, "get subscription")
	}

	return sub, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, sub domain.Subscription) (domain.Subscription, error) {
	const query = `
		UPDATE subscriptions
		SET service_name = $2,
			price = $3,
			user_id = $4,
			start_date = $5,
			end_date = $6
		WHERE id = $1
		RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at
	`

	updated, err := scanSubscription(r.pool.QueryRow(
		ctx,
		query,
		sub.ID,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate.Time(),
		monthPtrTime(sub.EndDate),
	))
	if err != nil {
		return domain.Subscription{}, mapNotFound(err, "update subscription")
	}

	return updated, nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `DELETE FROM subscriptions WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete subscription: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *SubscriptionRepository) List(ctx context.Context, filter domain.SubscriptionFilter) ([]domain.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
	`

	where, args := buildCommonFilters(filter.UserID, filter.ServiceName)
	query += where
	query += " ORDER BY created_at DESC, id DESC"

	if filter.Limit > 0 {
		args = append(args, filter.Limit)
		query += fmt.Sprintf(" LIMIT $%d", len(args))
	}

	if filter.Offset > 0 {
		args = append(args, filter.Offset)
		query += fmt.Sprintf(" OFFSET $%d", len(args))
	}

	subs, err := r.querySubscriptions(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}

	return subs, nil
}

func (r *SubscriptionRepository) ListForTotal(ctx context.Context, filter domain.TotalFilter) ([]domain.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
	`

	where, args := buildCommonFilters(filter.UserID, filter.ServiceName)
	conditions := make([]string, 0, 2)
	if where != "" {
		conditions = append(conditions, strings.TrimPrefix(where, " WHERE "))
	}

	args = append(args, filter.To.Time())
	conditions = append(conditions, fmt.Sprintf("start_date <= $%d", len(args)))

	args = append(args, filter.From.Time())
	conditions = append(conditions, fmt.Sprintf("(end_date IS NULL OR end_date >= $%d)", len(args)))

	query += " WHERE " + strings.Join(conditions, " AND ")
	query += " ORDER BY start_date ASC, id ASC"

	subs, err := r.querySubscriptions(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list subscriptions for total: %w", err)
	}

	return subs, nil
}

func (r *SubscriptionRepository) querySubscriptions(ctx context.Context, query string, args ...any) ([]domain.Subscription, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subs := make([]domain.Subscription, 0)
	for rows.Next() {
		sub, err := scanSubscription(rows)
		if err != nil {
			return nil, err
		}

		subs = append(subs, sub)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return subs, nil
}

type subscriptionScanner interface {
	Scan(dest ...any) error
}

func scanSubscription(scanner subscriptionScanner) (domain.Subscription, error) {
	var sub domain.Subscription
	var startDate time.Time
	var endDate *time.Time

	if err := scanner.Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&startDate,
		&endDate,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	); err != nil {
		return domain.Subscription{}, err
	}

	sub.StartDate = domain.NewMonth(startDate)
	if endDate != nil {
		endMonth := domain.NewMonth(*endDate)
		sub.EndDate = &endMonth
	}

	return sub, nil
}

func buildCommonFilters(userID *uuid.UUID, serviceName *string) (string, []any) {
	conditions := make([]string, 0, 2)
	args := make([]any, 0, 2)

	if userID != nil {
		args = append(args, *userID)
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", len(args)))
	}

	if serviceName != nil {
		args = append(args, *serviceName)
		conditions = append(conditions, fmt.Sprintf("service_name = $%d", len(args)))
	}

	if len(conditions) == 0 {
		return "", args
	}

	return " WHERE " + strings.Join(conditions, " AND "), args
}

func monthPtrTime(month *domain.Month) *time.Time {
	if month == nil {
		return nil
	}

	value := month.Time()
	return &value
}

func mapNotFound(err error, operation string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	return fmt.Errorf("%s: %w", operation, err)
}
