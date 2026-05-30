package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"subscription-service/internal/dto"
	"subscription-service/internal/service"
)

type SubscriptionHandler struct {
	svc    service.SubscriptionService
	logger *zap.Logger
}

func NewSubscriptionHandler(svc service.SubscriptionService, logger *zap.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		svc:    svc,
		logger: logger,
	}
}

// Create godoc
// @Summary      Create a subscription
// @Description  Create a new subscription record
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        subscription  body      dto.CreateSubscriptionRequest  true  "Subscription payload"
// @Success      201  {object}  dto.SubscriptionResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /subscriptions [post]
func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, h.logger, http.StatusBadRequest, "invalid request body")
		return
	}

	sub, err := req.ToDomain()
	if err != nil {
		writeError(w, h.logger, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.svc.Create(r.Context(), sub); err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeJSON(w, h.logger, http.StatusCreated, dto.NewSubscriptionResponse(sub))
}

// Get godoc
// @Summary      Get a subscription
// @Description  Get a subscription by its ID
// @Tags         subscriptions
// @Produce      json
// @Param        id   path      int  true  "Subscription ID"
// @Success      200  {object}  dto.SubscriptionResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Router       /subscriptions/{id} [get]
func (h *SubscriptionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, h.logger, http.StatusBadRequest, "invalid id")
		return
	}

	sub, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeJSON(w, h.logger, http.StatusOK, dto.NewSubscriptionResponse(sub))
}

// Update godoc
// @Summary      Update a subscription
// @Description  Update an existing subscription by its ID
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id            path      int                            true  "Subscription ID"
// @Param        subscription  body      dto.UpdateSubscriptionRequest  true  "Subscription payload"
// @Success      200  {object}  dto.SubscriptionResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /subscriptions/{id} [put]
func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, h.logger, http.StatusBadRequest, "invalid id")
		return
	}

	var req dto.UpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, h.logger, http.StatusBadRequest, "invalid request body")
		return
	}

	sub, err := req.ToDomain()
	if err != nil {
		writeError(w, h.logger, http.StatusBadRequest, err.Error())
		return
	}
	sub.ID = id

	if err := h.svc.Update(r.Context(), sub); err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeJSON(w, h.logger, http.StatusOK, dto.NewSubscriptionResponse(sub))
}

// Delete godoc
// @Summary      Delete a subscription
// @Description  Delete a subscription by its ID
// @Tags         subscriptions
// @Produce      json
// @Param        id   path      int  true  "Subscription ID"
// @Success      204  "No Content"
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Router       /subscriptions/{id} [delete]
func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, h.logger, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeJSON(w, h.logger, http.StatusNoContent, nil)
}

// List godoc
// @Summary      List subscriptions
// @Description  List subscriptions with optional filtering and pagination
// @Tags         subscriptions
// @Produce      json
// @Param        user_id       query     string  false  "Filter by user UUID"
// @Param        service_name  query     string  false  "Filter by service name"
// @Param        limit         query     int     false  "Page size"
// @Param        offset        query     int     false  "Page offset"
// @Success      200  {array}   dto.SubscriptionResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /subscriptions [get]
func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	params := service.ListParams{
		Limit:  parseInt(q.Get("limit"), 0),
		Offset: parseInt(q.Get("offset"), 0),
	}

	if raw := q.Get("user_id"); raw != "" {
		uid, err := uuid.Parse(raw)
		if err != nil {
			writeError(w, h.logger, http.StatusBadRequest, "invalid user_id")
			return
		}
		params.UserID = &uid
	}

	if name := q.Get("service_name"); name != "" {
		params.ServiceName = &name
	}

	subs, err := h.svc.List(r.Context(), params)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeJSON(w, h.logger, http.StatusOK, dto.NewSubscriptionListResponse(subs))
}

// TotalCost godoc
// @Summary      Total cost of subscriptions
// @Description  Calculate the total cost of subscriptions for a period with optional filtering by user and service name
// @Tags         subscriptions
// @Produce      json
// @Param        from          query     string  true   "Period start (MM-YYYY)"
// @Param        to            query     string  true   "Period end (MM-YYYY)"
// @Param        user_id       query     string  false  "Filter by user UUID"
// @Param        service_name  query     string  false  "Filter by service name"
// @Success      200  {object}  dto.TotalCostResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /subscriptions/cost [get]
func (h *SubscriptionHandler) TotalCost(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	from, err := time.Parse(dto.DateLayout, q.Get("from"))
	if err != nil {
		writeError(w, h.logger, http.StatusBadRequest, "invalid 'from': must be in format MM-YYYY")
		return
	}

	to, err := time.Parse(dto.DateLayout, q.Get("to"))
	if err != nil {
		writeError(w, h.logger, http.StatusBadRequest, "invalid 'to': must be in format MM-YYYY")
		return
	}

	params := service.CostParams{From: from, To: to}

	resp := dto.TotalCostResponse{
		From: q.Get("from"),
		To:   q.Get("to"),
	}

	if raw := q.Get("user_id"); raw != "" {
		uid, err := uuid.Parse(raw)
		if err != nil {
			writeError(w, h.logger, http.StatusBadRequest, "invalid user_id")
			return
		}
		params.UserID = &uid
		resp.UserID = &raw
	}

	if name := q.Get("service_name"); name != "" {
		params.ServiceName = &name
		resp.ServiceName = &name
	}

	total, err := h.svc.TotalCost(r.Context(), params)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	resp.Total = total
	writeJSON(w, h.logger, http.StatusOK, resp)
}

func (h *SubscriptionHandler) handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		writeError(w, h.logger, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrValidation):
		writeError(w, h.logger, http.StatusBadRequest, err.Error())
	default:
		writeError(w, h.logger, http.StatusInternalServerError, "internal server error")
	}
}

func parseID(r *http.Request) (uint, error) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func parseInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}
