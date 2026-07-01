package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"subscription-aggregator-service/internal/domain"
	"subscription-aggregator-service/internal/service"

	"github.com/go-chi/chi/v5"
)

type SubscriptionService interface {
	Create(ctx context.Context, input service.CreateSubscriptionInput) (domain.Subscription, error)
	GetByID(ctx context.Context, id string) (domain.Subscription, error)
	Update(ctx context.Context, input service.UpdateSubscriptionInput) (domain.Subscription, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, input service.ListSubscriptionsInput) ([]domain.Subscription, error)
	CalculateTotal(ctx context.Context, input service.CalculateTotalInput) (service.CalculateTotalResult, error)
}

type SubscriptionHandler struct {
	service SubscriptionService
	logger  *slog.Logger
}

func NewSubscriptionHandler(service SubscriptionService, logger *slog.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		service: service,
		logger:  logger,
	}
}

func (h *SubscriptionHandler) RegisterRoutes(r chi.Router) {
	r.Post("/subscriptions", h.Create)
	r.Get("/subscriptions", h.List)
	r.Get("/subscriptions/total", h.CalculateTotal)
	r.Get("/subscriptions/{id}", h.GetByID)
	r.Put("/subscriptions/{id}", h.Update)
	r.Delete("/subscriptions/{id}", h.Delete)
}

// Create godoc
// @Summary Create subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param request body CreateSubscriptionRequest true "Subscription payload"
// @Success 201 {object} SubscriptionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions [post]
func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateSubscriptionRequest
	if err := decodeJSON(r, &req); err != nil {
		h.logger.Warn("request_decode_failed", "operation", "create_subscription", "error", err)
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	sub, err := h.service.Create(r.Context(), service.CreateSubscriptionInput{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	})
	if err != nil {
		h.writeServiceError(w, "create_subscription_failed", err)
		return
	}

	h.logger.Info(
		"subscription_created",
		"subscription_id", sub.ID.String(),
		"user_id", sub.UserID.String(),
		"service_name", sub.ServiceName,
	)

	writeJSON(w, http.StatusCreated, NewSubscriptionResponse(sub))
}

// GetByID godoc
// @Summary Get subscription by ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} SubscriptionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	sub, err := h.service.GetByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, "get_subscription_failed", err)
		return
	}

	writeJSON(w, http.StatusOK, NewSubscriptionResponse(sub))
}

// Update godoc
// @Summary Update subscription
// @Description Performs full replacement of editable subscription fields.
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Param request body UpdateSubscriptionRequest true "Subscription payload"
// @Success 200 {object} SubscriptionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/{id} [put]
func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req UpdateSubscriptionRequest
	if err := decodeJSON(r, &req); err != nil {
		h.logger.Warn("request_decode_failed", "operation", "update_subscription", "error", err)
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	sub, err := h.service.Update(r.Context(), service.UpdateSubscriptionInput{
		ID:          chi.URLParam(r, "id"),
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	})
	if err != nil {
		h.writeServiceError(w, "update_subscription_failed", err)
		return
	}

	h.logger.Info(
		"subscription_updated",
		"subscription_id", sub.ID.String(),
		"user_id", sub.UserID.String(),
		"service_name", sub.ServiceName,
	)

	writeJSON(w, http.StatusOK, NewSubscriptionResponse(sub))
}

// Delete godoc
// @Summary Delete subscription
// @Tags subscriptions
// @Param id path string true "Subscription ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.service.Delete(r.Context(), id); err != nil {
		h.writeServiceError(w, "delete_subscription_failed", err)
		return
	}

	h.logger.Info("subscription_deleted", "subscription_id", id)

	w.WriteHeader(http.StatusNoContent)
}

// List godoc
// @Summary List subscriptions
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "User ID filter"
// @Param service_name query string false "Exact service name filter"
// @Param limit query int false "Pagination limit"
// @Param offset query int false "Pagination offset"
// @Success 200 {object} ListSubscriptionsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions [get]
func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, ok := parseIntQuery(w, r, "limit")
	if !ok {
		return
	}

	offset, ok := parseIntQuery(w, r, "offset")
	if !ok {
		return
	}

	subs, err := h.service.List(r.Context(), service.ListSubscriptionsInput{
		UserID:      r.URL.Query().Get("user_id"),
		ServiceName: r.URL.Query().Get("service_name"),
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		h.writeServiceError(w, "list_subscriptions_failed", err)
		return
	}

	h.logger.Info(
		"subscriptions_listed",
		"count", len(subs),
		"user_id", r.URL.Query().Get("user_id"),
		"service_name", r.URL.Query().Get("service_name"),
	)

	items := make([]SubscriptionResponse, 0, len(subs))
	for _, sub := range subs {
		items = append(items, NewSubscriptionResponse(sub))
	}

	writeJSON(w, http.StatusOK, ListSubscriptionsResponse{
		Items:  items,
		Limit:  limit,
		Offset: offset,
		Count:  len(items),
	})
}

// CalculateTotal godoc
// @Summary Calculate total subscription cost
// @Description Calculates total monthly cost for every active month in the inclusive period.
// @Tags subscriptions
// @Produce json
// @Param from query string true "Period start in MM-YYYY format"
// @Param to query string true "Period end in MM-YYYY format"
// @Param user_id query string false "User ID filter"
// @Param service_name query string false "Exact service name filter"
// @Success 200 {object} CalculateTotalResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/total [get]
func (h *SubscriptionHandler) CalculateTotal(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.CalculateTotal(r.Context(), service.CalculateTotalInput{
		From:        r.URL.Query().Get("from"),
		To:          r.URL.Query().Get("to"),
		UserID:      r.URL.Query().Get("user_id"),
		ServiceName: r.URL.Query().Get("service_name"),
	})
	if err != nil {
		h.writeServiceError(w, "calculate_total_failed", err)
		return
	}

	h.logger.Info(
		"subscription_total_calculated",
		"total", result.Total,
		"currency", result.Currency,
		"from", result.From.String(),
		"to", result.To.String(),
		"user_id", r.URL.Query().Get("user_id"),
		"service_name", r.URL.Query().Get("service_name"),
	)

	writeJSON(w, http.StatusOK, CalculateTotalResponse{
		Total:    result.Total,
		Currency: result.Currency,
		From:     result.From.String(),
		To:       result.To.String(),
	})
}

func (h *SubscriptionHandler) writeServiceError(w http.ResponseWriter, event string, err error) {
	switch {
	case errors.Is(err, service.ErrValidation):
		h.logger.Warn(event, "kind", "validation_error", "error", err)
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
	case errors.Is(err, service.ErrNotFound):
		h.logger.Warn(event, "kind", "not_found", "error", err)
		writeError(w, http.StatusNotFound, "not_found", err.Error())
	default:
		h.logger.Error(event, "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "unexpected server error")
	}
}

func decodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	return nil
}

func parseIntQuery(w http.ResponseWriter, r *http.Request, name string) (int, bool) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return 0, true
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", name+" must be an integer")
		return 0, false
	}

	return parsed, true
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, ErrorResponse{
		Error:   code,
		Message: message,
	})
}
