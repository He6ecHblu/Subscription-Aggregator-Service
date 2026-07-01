package handler

type CreateSubscriptionRequest struct {
	ServiceName string  `json:"service_name" binding:"required" example:"Yandex Plus"`
	Price       int     `json:"price" binding:"required" example:"400"`
	UserID      string  `json:"user_id" binding:"required" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string  `json:"start_date" binding:"required" example:"07-2025"`
	EndDate     *string `json:"end_date,omitempty" example:"12-2025"`
}

type UpdateSubscriptionRequest struct {
	ServiceName string  `json:"service_name" binding:"required" example:"Yandex Plus"`
	Price       int     `json:"price" binding:"required" example:"500"`
	UserID      string  `json:"user_id" binding:"required" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string  `json:"start_date" binding:"required" example:"07-2025"`
	EndDate     *string `json:"end_date,omitempty" example:"12-2025"`
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
