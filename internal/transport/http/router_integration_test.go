//go:build integration

package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"subscription-service/internal/domain"
	"subscription-service/internal/dto"
	"subscription-service/internal/repository"
	"subscription-service/internal/service"
	transport "subscription-service/internal/transport/http"
)

type memRepo struct {
	items  map[uint]domain.Subscription
	nextID uint
}

func newMemRepo() *memRepo {
	return &memRepo{items: make(map[uint]domain.Subscription), nextID: 1}
}

func (m *memRepo) Create(_ context.Context, sub *domain.Subscription) error {
	sub.ID = m.nextID
	m.nextID++
	m.items[sub.ID] = *sub
	return nil
}

func (m *memRepo) Get(_ context.Context, id uint) (*domain.Subscription, error) {
	sub, ok := m.items[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return &sub, nil
}

func (m *memRepo) Update(_ context.Context, sub *domain.Subscription) error {
	if _, ok := m.items[sub.ID]; !ok {
		return repository.ErrNotFound
	}
	m.items[sub.ID] = *sub
	return nil
}

func (m *memRepo) Delete(_ context.Context, id uint) error {
	if _, ok := m.items[id]; !ok {
		return repository.ErrNotFound
	}
	delete(m.items, id)
	return nil
}

func (m *memRepo) List(_ context.Context, _ repository.ListFilter) ([]domain.Subscription, error) {
	out := make([]domain.Subscription, 0, len(m.items))
	for _, v := range m.items {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (m *memRepo) TotalCost(_ context.Context, f repository.CostFilter) (int64, error) {
	var total int64
	for _, v := range m.items {
		if f.UserID != nil && v.UserID != *f.UserID {
			continue
		}
		if f.ServiceName != nil && v.ServiceName != *f.ServiceName {
			continue
		}
		total += int64(v.Price)
	}
	return total, nil
}

func newTestServer() *httptest.Server {
	logger := zap.NewNop()
	svc := service.NewSubscriptionService(newMemRepo(), logger)
	handler := transport.NewSubscriptionHandler(svc, logger)
	return httptest.NewServer(transport.NewRouter(handler, logger))
}

func TestIntegration_CreateAndGet(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	body := dto.CreateSubscriptionRequest{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      uuid.New().String(),
		StartDate:   "07-2025",
	}
	raw, _ := json.Marshal(body)

	resp, err := http.Post(srv.URL+"/subscriptions", "application/json", bytes.NewReader(raw))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created dto.SubscriptionResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	resp.Body.Close()

	assert.Equal(t, uint(1), created.ID)
	assert.Equal(t, "Yandex Plus", created.ServiceName)

	getResp, err := http.Get(srv.URL + "/subscriptions/1")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, getResp.StatusCode)

	var got dto.SubscriptionResponse
	require.NoError(t, json.NewDecoder(getResp.Body).Decode(&got))
	getResp.Body.Close()

	assert.Equal(t, created, got)
}

func TestIntegration_TotalCost(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	userID := uuid.New().String()
	prices := []int{400, 600}
	for _, p := range prices {
		body := dto.CreateSubscriptionRequest{
			ServiceName: "Netflix",
			Price:       p,
			UserID:      userID,
			StartDate:   "01-2025",
		}
		raw, _ := json.Marshal(body)
		resp, err := http.Post(srv.URL+"/subscriptions", "application/json", bytes.NewReader(raw))
		require.NoError(t, err)
		resp.Body.Close()
	}

	costResp, err := http.Get(srv.URL + "/subscriptions/cost?from=01-2025&to=12-2025&user_id=" + userID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, costResp.StatusCode)

	var cost dto.TotalCostResponse
	require.NoError(t, json.NewDecoder(costResp.Body).Decode(&cost))
	costResp.Body.Close()

	assert.Equal(t, int64(1000), cost.Total)
}
