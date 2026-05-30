package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"subscription-service/internal/domain"
	"subscription-service/internal/repository"
)

type fakeRepo struct {
	createFn    func(ctx context.Context, sub *domain.Subscription) error
	getFn       func(ctx context.Context, id uint) (*domain.Subscription, error)
	updateFn    func(ctx context.Context, sub *domain.Subscription) error
	deleteFn    func(ctx context.Context, id uint) error
	listFn      func(ctx context.Context, f repository.ListFilter) ([]domain.Subscription, error)
	totalCostFn func(ctx context.Context, f repository.CostFilter) (int64, error)
}

func (f *fakeRepo) Create(ctx context.Context, sub *domain.Subscription) error {
	return f.createFn(ctx, sub)
}

func (f *fakeRepo) Get(ctx context.Context, id uint) (*domain.Subscription, error) {
	return f.getFn(ctx, id)
}

func (f *fakeRepo) Update(ctx context.Context, sub *domain.Subscription) error {
	return f.updateFn(ctx, sub)
}

func (f *fakeRepo) Delete(ctx context.Context, id uint) error {
	return f.deleteFn(ctx, id)
}

func (f *fakeRepo) List(ctx context.Context, filter repository.ListFilter) ([]domain.Subscription, error) {
	return f.listFn(ctx, filter)
}

func (f *fakeRepo) TotalCost(ctx context.Context, filter repository.CostFilter) (int64, error) {
	return f.totalCostFn(ctx, filter)
}

func newValidSubscription() *domain.Subscription {
	return &domain.Subscription{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      uuid.New(),
		StartDate:   time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
	}
}

func TestCreate_Success(t *testing.T) {
	called := false
	repo := &fakeRepo{createFn: func(_ context.Context, sub *domain.Subscription) error {
		called = true
		sub.ID = 1
		return nil
	}}
	svc := NewSubscriptionService(repo, zap.NewNop())

	err := svc.Create(context.Background(), newValidSubscription())

	require.NoError(t, err)
	assert.True(t, called)
}

func TestCreate_EmptyServiceName_ReturnsValidationError(t *testing.T) {
	repo := &fakeRepo{createFn: func(_ context.Context, _ *domain.Subscription) error {
		t.Fatal("repo.Create must not be called on invalid input")
		return nil
	}}
	svc := NewSubscriptionService(repo, zap.NewNop())

	sub := newValidSubscription()
	sub.ServiceName = ""

	err := svc.Create(context.Background(), sub)

	assert.ErrorIs(t, err, ErrValidation)
}

func TestCreate_EndBeforeStart_ReturnsValidationError(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewSubscriptionService(repo, zap.NewNop())

	sub := newValidSubscription()
	end := sub.StartDate.AddDate(0, -1, 0)
	sub.EndDate = &end

	err := svc.Create(context.Background(), sub)

	assert.ErrorIs(t, err, ErrValidation)
}

func TestGet_NotFound_MapsToServiceError(t *testing.T) {
	repo := &fakeRepo{getFn: func(_ context.Context, _ uint) (*domain.Subscription, error) {
		return nil, repository.ErrNotFound
	}}
	svc := NewSubscriptionService(repo, zap.NewNop())

	_, err := svc.Get(context.Background(), 42)

	assert.ErrorIs(t, err, ErrNotFound)
}

func TestTotalCost_ToBeforeFrom_ReturnsValidationError(t *testing.T) {
	repo := &fakeRepo{totalCostFn: func(_ context.Context, _ repository.CostFilter) (int64, error) {
		t.Fatal("repo.TotalCost must not be called on invalid period")
		return 0, nil
	}}
	svc := NewSubscriptionService(repo, zap.NewNop())

	params := CostParams{
		From: time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	_, err := svc.TotalCost(context.Background(), params)

	assert.ErrorIs(t, err, ErrValidation)
}

func TestTotalCost_Success_PassesFilterThrough(t *testing.T) {
	var captured repository.CostFilter
	repo := &fakeRepo{totalCostFn: func(_ context.Context, f repository.CostFilter) (int64, error) {
		captured = f
		return 1200, nil
	}}
	svc := NewSubscriptionService(repo, zap.NewNop())

	uid := uuid.New()
	params := CostParams{
		From:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		To:     time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
		UserID: &uid,
	}

	total, err := svc.TotalCost(context.Background(), params)

	require.NoError(t, err)
	assert.Equal(t, int64(1200), total)
	assert.Equal(t, &uid, captured.UserID)
}
