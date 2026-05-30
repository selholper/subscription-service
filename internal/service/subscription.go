package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"subscription-service/internal/domain"
)

//go:generate mockgen -source=subscription.go -destination=mocks/subscription_mock.go -package=mocks

type SubscriptionService interface {
	Create(ctx context.Context, sub *domain.Subscription) error
	Get(ctx context.Context, id uint) (*domain.Subscription, error)
	Update(ctx context.Context, sub *domain.Subscription) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, params ListParams) ([]domain.Subscription, error)
	TotalCost(ctx context.Context, params CostParams) (int64, error)
}

type ListParams struct {
	Limit       int
	Offset      int
	UserID      *uuid.UUID
	ServiceName *string
}

type CostParams struct {
	From        time.Time
	To          time.Time
	UserID      *uuid.UUID
	ServiceName *string
}

var (
	ErrNotFound   = errors.New("subscription not found")
	ErrValidation = errors.New("validation error")
)
