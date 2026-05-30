package service

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"subscription-service/internal/domain"
	"subscription-service/internal/repository"
)

type subscriptionService struct {
	repo   repository.SubscriptionRepository
	logger *zap.Logger
}

func NewSubscriptionService(
	repo repository.SubscriptionRepository,
	logger *zap.Logger,
) SubscriptionService {
	return &subscriptionService{
		repo:   repo,
		logger: logger,
	}
}

func (s *subscriptionService) Create(ctx context.Context, sub *domain.Subscription) error {
	if err := validateSubscription(sub); err != nil {
		s.logger.Warn("create subscription validation failed", zap.Error(err))
		return err
	}

	if err := s.repo.Create(ctx, sub); err != nil {
		s.logger.Error("failed to create subscription", zap.Error(err))
		return err
	}

	s.logger.Info("subscription created",
		zap.Uint("id", sub.ID),
		zap.String("service_name", sub.ServiceName),
		zap.String("user_id", sub.UserID.String()),
	)

	return nil
}

func (s *subscriptionService) Get(ctx context.Context, id uint) (*domain.Subscription, error) {
	sub, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		s.logger.Error("failed to get subscription", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}

	return sub, nil
}

func (s *subscriptionService) Update(ctx context.Context, sub *domain.Subscription) error {
	if err := validateSubscription(sub); err != nil {
		s.logger.Warn("update subscription validation failed", zap.Error(err))
		return err
	}

	if err := s.repo.Update(ctx, sub); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		s.logger.Error("failed to update subscription", zap.Uint("id", sub.ID), zap.Error(err))
		return err
	}

	s.logger.Info("subscription updated", zap.Uint("id", sub.ID))

	return nil
}

func (s *subscriptionService) Delete(ctx context.Context, id uint) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		s.logger.Error("failed to delete subscription", zap.Uint("id", id), zap.Error(err))
		return err
	}

	s.logger.Info("subscription deleted", zap.Uint("id", id))

	return nil
}

func (s *subscriptionService) List(ctx context.Context, params ListParams) ([]domain.Subscription, error) {
	subs, err := s.repo.List(ctx, repository.ListFilter{
		Limit:       params.Limit,
		Offset:      params.Offset,
		UserID:      params.UserID,
		ServiceName: params.ServiceName,
	})
	if err != nil {
		s.logger.Error("failed to list subscriptions", zap.Error(err))
		return nil, err
	}

	s.logger.Info("subscriptions listed", zap.Int("count", len(subs)))

	return subs, nil
}

func (s *subscriptionService) TotalCost(ctx context.Context, params CostParams) (int64, error) {
	if params.To.Before(params.From) {
		return 0, fmt.Errorf("%w: 'to' must not be before 'from'", ErrValidation)
	}

	total, err := s.repo.TotalCost(ctx, repository.CostFilter{
		PeriodStart: params.From,
		PeriodEnd:   params.To,
		UserID:      params.UserID,
		ServiceName: params.ServiceName,
	})
	if err != nil {
		s.logger.Error("failed to calculate total cost", zap.Error(err))
		return 0, err
	}

	s.logger.Info("total cost calculated",
		zap.Int64("total", total),
		zap.Time("from", params.From),
		zap.Time("to", params.To),
	)

	return total, nil
}

func validateSubscription(sub *domain.Subscription) error {
	if sub.ServiceName == "" {
		return fmt.Errorf("%w: service_name is required", ErrValidation)
	}
	if sub.Price < 0 {
		return fmt.Errorf("%w: price must not be negative", ErrValidation)
	}
	if sub.EndDate != nil && sub.EndDate.Before(sub.StartDate) {
		return fmt.Errorf("%w: end_date must not be before start_date", ErrValidation)
	}

	return nil
}
