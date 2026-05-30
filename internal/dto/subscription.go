package dto

import (
	"time"

	"github.com/google/uuid"

	"subscription-service/internal/domain"
)

const DateLayout = "01-2006"

type CreateSubscriptionRequest struct {
	ServiceName string  `json:"service_name" example:"Yandex Plus"`
	Price       int     `json:"price" example:"400"`
	UserID      string  `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string  `json:"start_date" example:"07-2025"`
	EndDate     *string `json:"end_date,omitempty" example:"12-2025"`
}

type UpdateSubscriptionRequest struct {
	ServiceName string  `json:"service_name" example:"Yandex Plus"`
	Price       int     `json:"price" example:"500"`
	UserID      string  `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string  `json:"start_date" example:"07-2025"`
	EndDate     *string `json:"end_date,omitempty" example:"12-2025"`
}

type SubscriptionResponse struct {
	ID          uint    `json:"id" example:"1"`
	ServiceName string  `json:"service_name" example:"Yandex Plus"`
	Price       int     `json:"price" example:"400"`
	UserID      string  `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string  `json:"start_date" example:"07-2025"`
	EndDate     *string `json:"end_date,omitempty" example:"12-2025"`
}

type TotalCostResponse struct {
	Total       int64   `json:"total" example:"1200"`
	From        string  `json:"from" example:"01-2025"`
	To          string  `json:"to" example:"12-2025"`
	UserID      *string `json:"user_id,omitempty" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	ServiceName *string `json:"service_name,omitempty" example:"Yandex Plus"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"invalid request body"`
}

func (r CreateSubscriptionRequest) ToDomain() (*domain.Subscription, error) {
	return buildSubscription(r.ServiceName, r.Price, r.UserID, r.StartDate, r.EndDate)
}

func (r UpdateSubscriptionRequest) ToDomain() (*domain.Subscription, error) {
	return buildSubscription(r.ServiceName, r.Price, r.UserID, r.StartDate, r.EndDate)
}

func buildSubscription(
	serviceName string,
	price int,
	userID string,
	startDate string,
	endDate *string,
) (*domain.Subscription, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	var start time.Time
	start, err = time.Parse(DateLayout, startDate)
	if err != nil {
		return nil, ErrInvalidStartDate
	}

	var end *time.Time
	if endDate != nil && *endDate != "" {
		var parsed time.Time
		parsed, err = time.Parse(DateLayout, *endDate)
		if err != nil {
			return nil, ErrInvalidEndDate
		}
		end = &parsed
	}

	return &domain.Subscription{
		ServiceName: serviceName,
		Price:       price,
		UserID:      uid,
		StartDate:   start,
		EndDate:     end,
	}, nil
}

func NewSubscriptionResponse(sub *domain.Subscription) SubscriptionResponse {
	resp := SubscriptionResponse{
		ID:          sub.ID,
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		UserID:      sub.UserID.String(),
		StartDate:   sub.StartDate.Format(DateLayout),
	}

	if sub.EndDate != nil {
		end := sub.EndDate.Format(DateLayout)
		resp.EndDate = &end
	}

	return resp
}

func NewSubscriptionListResponse(subs []domain.Subscription) []SubscriptionResponse {
	out := make([]SubscriptionResponse, 0, len(subs))
	for i := range subs {
		out = append(out, NewSubscriptionResponse(&subs[i]))
	}
	return out
}
