package domain

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID
	ServiceName string
	Price       int
	UserID      uuid.UUID
	StartDate   Month
	EndDate     *Month
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type SubscriptionFilter struct {
	UserID      *uuid.UUID
	ServiceName *string
	Limit       int
	Offset      int
}

type TotalFilter struct {
	From        Month
	To          Month
	UserID      *uuid.UUID
	ServiceName *string
}
