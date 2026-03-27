package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"laundry-backend/internal/model"
	"laundry-backend/internal/repository"
)

type fakeOrderRepo struct {
	createIn repository.CreateOrderParams
}

func (f *fakeOrderRepo) List(ctx context.Context, q string, page, pageSize int, sort string, dir string, startDate, endDate *time.Time) (model.Paged[model.OrderListItem], error) {
	return model.Paged[model.OrderListItem]{}, nil
}
func (f *fakeOrderRepo) GetDetail(ctx context.Context, id string) (*model.OrderDetail, error) {
	return nil, nil
}
func (f *fakeOrderRepo) GetDetailByPublicToken(ctx context.Context, token string) (*model.OrderDetail, error) {
	return nil, nil
}
func (f *fakeOrderRepo) Create(ctx context.Context, p repository.CreateOrderParams) (*model.OrderDetail, error) {
	f.createIn = p
	return &model.OrderDetail{ID: "x"}, nil
}
func (f *fakeOrderRepo) UpdateImage(ctx context.Context, orderID string, image *string) error {
	return nil
}
func (f *fakeOrderRepo) UpdateWorkflow(ctx context.Context, orderID string, workflowStatus string) error {
	return nil
}
func (f *fakeOrderRepo) CreatePayment(ctx context.Context, orderID string, p repository.CreatePaymentParams) (*model.Payment, error) {
	return nil, nil
}
func (f *fakeOrderRepo) UpsertWorkAssignment(ctx context.Context, p repository.UpsertWorkAssignmentParams) error {
	return nil
}
func (f *fakeOrderRepo) DeleteWorkAssignment(ctx context.Context, orderItemID string, taskType string) error {
	return nil
}
func (f *fakeOrderRepo) CreateAttachments(ctx context.Context, orderID string, files []repository.CreateAttachmentParams) error {
	return nil
}
func (f *fakeOrderRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func TestOrderServiceCreate_ComputesTotals(t *testing.T) {
	repo := &fakeOrderRepo{}
	svc := NewOrderService(repo)

	now := time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC)
	_, err := svc.Create(context.Background(), CreateOrderInput{
		CustomerID:   "cust1",
		ReceivedDate: now,
		Items: []CreateOrderItemInput{
			{ServiceTypeID: "svc1", Quantity: 2, UnitPrice: 10000, Discount: 1000},
			{ServiceTypeID: "svc2", Quantity: 1.5, UnitPrice: 2000.5, Discount: 0},
		},
	})
	require.NoError(t, err)

	require.Equal(t, "cust1", repo.createIn.CustomerID)
	require.Equal(t, now, repo.createIn.ReceivedDate)
	require.Len(t, repo.createIn.Items, 2)

	require.Equal(t, "2.00", repo.createIn.Items[0].Quantity)
	require.Equal(t, "10000.00", repo.createIn.Items[0].UnitPrice)
	require.Equal(t, "1000.00", repo.createIn.Items[0].Discount)
	require.Equal(t, "19000.00", repo.createIn.Items[0].Total)

	require.Equal(t, "1.50", repo.createIn.Items[1].Quantity)
	require.Equal(t, "2000.50", repo.createIn.Items[1].UnitPrice)
	require.Equal(t, "0.00", repo.createIn.Items[1].Discount)
	require.Equal(t, "3000.75", repo.createIn.Items[1].Total)
}
