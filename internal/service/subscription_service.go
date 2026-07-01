package service

import (
	"context"
	"errors"
	"strings"

	"subscription-aggregator-service/internal/domain"
	"subscription-aggregator-service/internal/repository"

	"github.com/google/uuid"
)

const defaultCurrency = "RUB"

type SubscriptionRepository interface {
	Create(ctx context.Context, sub domain.Subscription) (domain.Subscription, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Subscription, error)
	Update(ctx context.Context, sub domain.Subscription) (domain.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.SubscriptionFilter) ([]domain.Subscription, error)
	ListForTotal(ctx context.Context, filter domain.TotalFilter) ([]domain.Subscription, error)
}

type SubscriptionService struct {
	repo SubscriptionRepository
}

func NewSubscriptionService(repo SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

type CreateSubscriptionInput struct {
	ServiceName string
	Price       int
	UserID      string
	StartDate   string
	EndDate     *string
}

type UpdateSubscriptionInput struct {
	ID          string
	ServiceName string
	Price       int
	UserID      string
	StartDate   string
	EndDate     *string
}

type ListSubscriptionsInput struct {
	UserID      string
	ServiceName string
	Limit       int
	Offset      int
}

type CalculateTotalInput struct {
	From        string
	To          string
	UserID      string
	ServiceName string
}

type CalculateTotalResult struct {
	Total    int
	Currency string
	From     domain.Month
	To       domain.Month
}

func (s *SubscriptionService) Create(ctx context.Context, input CreateSubscriptionInput) (domain.Subscription, error) {
	sub, err := parseSubscriptionInput(input.ServiceName, input.Price, input.UserID, input.StartDate, input.EndDate)
	if err != nil {
		return domain.Subscription{}, err
	}

	created, err := s.repo.Create(ctx, sub)
	if err != nil {
		return domain.Subscription{}, wrapError("create subscription", err)
	}

	return created, nil
}

func (s *SubscriptionService) GetByID(ctx context.Context, id string) (domain.Subscription, error) {
	subscriptionID, err := parseUUID("id", id)
	if err != nil {
		return domain.Subscription{}, err
	}

	sub, err := s.repo.GetByID(ctx, subscriptionID)
	if err != nil {
		return domain.Subscription{}, mapRepositoryError(err, "subscription not found")
	}

	return sub, nil
}

func (s *SubscriptionService) Update(ctx context.Context, input UpdateSubscriptionInput) (domain.Subscription, error) {
	subscriptionID, err := parseUUID("id", input.ID)
	if err != nil {
		return domain.Subscription{}, err
	}

	sub, err := parseSubscriptionInput(input.ServiceName, input.Price, input.UserID, input.StartDate, input.EndDate)
	if err != nil {
		return domain.Subscription{}, err
	}
	sub.ID = subscriptionID

	updated, err := s.repo.Update(ctx, sub)
	if err != nil {
		return domain.Subscription{}, mapRepositoryError(err, "subscription not found")
	}

	return updated, nil
}

func (s *SubscriptionService) Delete(ctx context.Context, id string) error {
	subscriptionID, err := parseUUID("id", id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, subscriptionID); err != nil {
		return mapRepositoryError(err, "subscription not found")
	}

	return nil
}

func (s *SubscriptionService) List(ctx context.Context, input ListSubscriptionsInput) ([]domain.Subscription, error) {
	if input.Limit < 0 {
		return nil, validationError("limit must not be negative")
	}

	if input.Offset < 0 {
		return nil, validationError("offset must not be negative")
	}

	userID, err := parseOptionalUUID("user_id", input.UserID)
	if err != nil {
		return nil, err
	}

	serviceName := parseOptionalServiceName(input.ServiceName)

	subs, err := s.repo.List(ctx, domain.SubscriptionFilter{
		UserID:      userID,
		ServiceName: serviceName,
		Limit:       input.Limit,
		Offset:      input.Offset,
	})
	if err != nil {
		return nil, wrapError("list subscriptions", err)
	}

	return subs, nil
}

func (s *SubscriptionService) CalculateTotal(ctx context.Context, input CalculateTotalInput) (CalculateTotalResult, error) {
	from, err := parseMonthField("from", input.From)
	if err != nil {
		return CalculateTotalResult{}, err
	}

	to, err := parseMonthField("to", input.To)
	if err != nil {
		return CalculateTotalResult{}, err
	}

	if from.After(to) {
		return CalculateTotalResult{}, validationError("from must not be after to")
	}

	userID, err := parseOptionalUUID("user_id", input.UserID)
	if err != nil {
		return CalculateTotalResult{}, err
	}

	serviceName := parseOptionalServiceName(input.ServiceName)

	filter := domain.TotalFilter{
		From:        from,
		To:          to,
		UserID:      userID,
		ServiceName: serviceName,
	}

	subs, err := s.repo.ListForTotal(ctx, filter)
	if err != nil {
		return CalculateTotalResult{}, wrapError("list subscriptions for total", err)
	}

	total, err := calculateTotal(subs, from, to)
	if err != nil {
		return CalculateTotalResult{}, err
	}

	return CalculateTotalResult{
		Total:    total,
		Currency: defaultCurrency,
		From:     from,
		To:       to,
	}, nil
}

func parseSubscriptionInput(serviceName string, price int, userIDValue string, startDate string, endDateValue *string) (domain.Subscription, error) {
	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		return domain.Subscription{}, validationError("service_name must not be empty")
	}

	if price <= 0 {
		return domain.Subscription{}, validationError("price must be positive")
	}

	userID, err := parseUUID("user_id", userIDValue)
	if err != nil {
		return domain.Subscription{}, err
	}

	start, err := parseMonthField("start_date", startDate)
	if err != nil {
		return domain.Subscription{}, err
	}

	var end *domain.Month
	if endDateValue != nil {
		parsedEnd, err := parseMonthField("end_date", *endDateValue)
		if err != nil {
			return domain.Subscription{}, err
		}

		if parsedEnd.Before(start) {
			return domain.Subscription{}, validationError("end_date must not be earlier than start_date")
		}

		end = &parsedEnd
	}

	return domain.Subscription{
		ServiceName: serviceName,
		Price:       price,
		UserID:      userID,
		StartDate:   start,
		EndDate:     end,
	}, nil
}

func calculateTotal(subs []domain.Subscription, from domain.Month, to domain.Month) (int, error) {
	total := 0

	for _, sub := range subs {
		activeFrom := maxMonth(sub.StartDate, from)
		activeTo := to
		if sub.EndDate != nil {
			activeTo = minMonth(*sub.EndDate, to)
		}

		if activeFrom.After(activeTo) {
			continue
		}

		months, err := activeFrom.MonthsUntilInclusive(activeTo)
		if err != nil {
			return 0, validationError(err.Error())
		}

		total += sub.Price * months
	}

	return total, nil
}

func parseUUID(field string, value string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		return uuid.Nil, validationError(field + " must be a valid UUID")
	}

	return parsed, nil
}

func parseOptionalUUID(field string, value string) (*uuid.UUID, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parsed, err := parseUUID(field, value)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}

func parseMonthField(field string, value string) (domain.Month, error) {
	month, err := domain.ParseMonth(strings.TrimSpace(value))
	if err != nil {
		return domain.Month{}, validationError(field + " must have MM-YYYY format")
	}

	return month, nil
}

func parseOptionalServiceName(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return &value
}

func maxMonth(left domain.Month, right domain.Month) domain.Month {
	if left.After(right) {
		return left
	}

	return right
}

func minMonth(left domain.Month, right domain.Month) domain.Month {
	if left.Before(right) {
		return left
	}

	return right
}

func mapRepositoryError(err error, notFoundMessage string) error {
	if errors.Is(err, repository.ErrNotFound) {
		return notFoundError(notFoundMessage)
	}

	return err
}
