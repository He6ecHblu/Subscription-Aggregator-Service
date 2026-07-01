package handler

type CreateSubscriptionRequest struct {
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date,omitempty"`
}

type UpdateSubscriptionRequest struct {
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date,omitempty"`
}

type SubscriptionResponse struct {
	ID          string  `json:"id"`
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type ListSubscriptionsResponse struct {
	Items  []SubscriptionResponse `json:"items"`
	Limit  int                    `json:"limit"`
	Offset int                    `json:"offset"`
	Count  int                    `json:"count"`
}

type CalculateTotalResponse struct {
	Total    int    `json:"total"`
	Currency string `json:"currency"`
	From     string `json:"from"`
	To       string `json:"to"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
