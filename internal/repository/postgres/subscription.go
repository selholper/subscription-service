package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"subscription-service/internal/domain"
	"subscription-service/internal/repository"
)

type subscriptionRepo struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) repository.SubscriptionRepository {
	return &subscriptionRepo{db: db}
}

func (r *subscriptionRepo) Create(
	ctx context.Context,
	sub *domain.Subscription,
) error {
	return r.db.WithContext(ctx).Create(sub).Error
}

func (r *subscriptionRepo) Get(
	ctx context.Context,
	id uint,
) (*domain.Subscription, error) {
	var sub domain.Subscription

	err := r.db.WithContext(ctx).First(&sub, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return &sub, nil
}

func (r *subscriptionRepo) Update(
	ctx context.Context,
	sub *domain.Subscription,
) error {
	result := r.db.WithContext(ctx).Save(sub)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *subscriptionRepo) Delete(
	ctx context.Context,
	id uint,
) error {
	result := r.db.WithContext(ctx).Delete(&domain.Subscription{}, id)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *subscriptionRepo) List(
	ctx context.Context,
	filter repository.ListFilter,
) ([]domain.Subscription, error) {
	var subs []domain.Subscription

	q := r.db.WithContext(ctx).Model(&domain.Subscription{})

	if filter.UserID != nil {
		q = q.Where("user_id = ?", *filter.UserID)
	}
	if filter.ServiceName != nil {
		q = q.Where("service_name = ?", *filter.ServiceName)
	}
	if filter.Limit > 0 {
		q = q.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		q = q.Offset(filter.Offset)
	}

	q = q.Order("id ASC")

	if err := q.Find(&subs).Error; err != nil {
		return nil, err
	}

	return subs, nil
}

func periodFilter(q *gorm.DB, f repository.CostFilter) *gorm.DB {
	return q.
		Where("start_date <= ?", f.PeriodEnd).
		Where("end_date IS NULL OR end_date >= ?", f.PeriodStart)
}

func (r *subscriptionRepo) TotalCost(
	ctx context.Context,
	filter repository.CostFilter,
) (int64, error) {
	var total int64

	q := r.db.WithContext(ctx).
		Model(&domain.Subscription{}).
		Select("COALESCE(SUM(price), 0)")

	q = periodFilter(q, filter)

	if filter.UserID != nil {
		q = q.Where("user_id = ?", *filter.UserID)
	}
	if filter.ServiceName != nil {
		q = q.Where("service_name = ?", *filter.ServiceName)
	}

	if err := q.Scan(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}
