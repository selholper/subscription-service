package repository

import (
	"context"
	"subscription-service/internal/domain"
	"time"

	"github.com/google/uuid"
)

//go:generate mockgen -source=subscription.go -destination=mocks/subscription_mock.go -package=mocks

type SubscriptionRepository interface {
	Create(
		ctx context.Context,
		sub *domain.Subscription,
	) error

	Get(
		ctx context.Context,
		id uint,
	) (*domain.Subscription, error)

	Update(
		ctx context.Context,
		sub *domain.Subscription,
	) error

	Delete(
		ctx context.Context,
		id uint,
	) error

	List(
		ctx context.Context,
		filter ListFilter,
	) ([]domain.Subscription, error)

	TotalCost(
		ctx context.Context,
		filter CostFilter,
	) (int64, error)
}

type ListFilter struct {
	Limit       int
	Offset      int
	UserID      *uuid.UUID
	ServiceName *string
}

type CostFilter struct {
	PeriodStart time.Time
	PeriodEnd   time.Time
	UserID      *uuid.UUID
	ServiceName *string
}

type repoError struct{ msg string }

func (e *repoError) Error() string { return "subscription repository: " + e.msg }

var ErrNotFound = &repoError{"not found"}
