package handler

import (
	"time"

	"subscription-aggregator-service/internal/domain"
)

func NewSubscriptionResponse(sub domain.Subscription) SubscriptionResponse {
	var endDate *string
	if sub.EndDate != nil {
		formatted := sub.EndDate.String()
		endDate = &formatted
	}

	return SubscriptionResponse{
		ID:          sub.ID.String(),
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		UserID:      sub.UserID.String(),
		StartDate:   sub.StartDate.String(),
		EndDate:     endDate,
		CreatedAt:   sub.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   sub.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
