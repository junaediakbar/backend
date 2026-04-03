package repository

import (
	"context"
	"errors"
	"time"

	"laundry-backend/internal/model"
)

var (
	ErrCustomerHasOrders         = errors.New("customer has orders")
	ErrCustomerHasDeliveryStops  = errors.New("customer has delivery stops")
)

type CustomerRepository interface {
	List(ctx context.Context, q string, page, pageSize int) (model.Paged[model.Customer], error)
	Get(ctx context.Context, id string) (*model.Customer, error)
	RecentOrders(ctx context.Context, customerID string, limit int) ([]model.CustomerOrderSummary, error)
	Create(ctx context.Context, c CreateCustomerParams) (*model.Customer, error)
	Update(ctx context.Context, id string, c UpdateCustomerParams) (*model.Customer, error)
	Delete(ctx context.Context, id string) error
}

type CreateCustomerParams struct {
	Name      string
	Phone     *string
	Address   *string
	Latitude  *float64
	Longitude *float64
	Email     *string
	Notes     *string
}

type UpdateCustomerParams = CreateCustomerParams

type ServiceTypeRepository interface {
	List(ctx context.Context, onlyActive *bool) ([]model.ServiceType, error)
	Get(ctx context.Context, id string) (*model.ServiceType, error)
	Create(ctx context.Context, p CreateServiceTypeParams) (*model.ServiceType, error)
	Update(ctx context.Context, id string, p UpdateServiceTypeParams) (*model.ServiceType, error)
	Delete(ctx context.Context, id string) error
}

type CreateServiceTypeParams struct {
	Name         string
	Unit         string
	DefaultPrice string
	IsActive     bool
}

type UpdateServiceTypeParams = CreateServiceTypeParams

type EmployeeRepository interface {
	List(ctx context.Context, onlyActive *bool) ([]model.Employee, error)
	Get(ctx context.Context, id string) (*model.Employee, error)
	Create(ctx context.Context, p CreateEmployeeParams) (*model.Employee, error)
	Update(ctx context.Context, id string, p UpdateEmployeeParams) (*model.Employee, error)
	Delete(ctx context.Context, id string) error
	Performance(ctx context.Context, start, end *time.Time) ([]model.EmployeePerformanceRow, error)
}

type CreateEmployeeParams struct {
	Name     string
	IsActive bool
}

type UpdateEmployeeParams = CreateEmployeeParams

type OrderRepository interface {
	List(ctx context.Context, q string, page, pageSize int, sort string, dir string, startDate, endDate *time.Time) (model.Paged[model.OrderListItem], error)
	GetDetail(ctx context.Context, id string) (*model.OrderDetail, error)
	GetDetailByPublicToken(ctx context.Context, token string) (*model.OrderDetail, error)
	Create(ctx context.Context, p CreateOrderParams) (*model.OrderDetail, error)
	Delete(ctx context.Context, id string) error
	UpdateImage(ctx context.Context, orderID string, image *string) error
	UpdateWorkflow(ctx context.Context, orderID string, workflowStatus string) error
	CreatePayment(ctx context.Context, orderID string, p CreatePaymentParams) (*model.Payment, error)
	DeletePayment(ctx context.Context, orderID string, paymentID string) (*model.Payment, error)
	UpsertWorkAssignment(ctx context.Context, p UpsertWorkAssignmentParams) error
	DeleteWorkAssignment(ctx context.Context, orderItemID string, taskType string) error
	CreateAttachments(ctx context.Context, orderID string, files []CreateAttachmentParams) error
}

type CreateOrderParams struct {
	CustomerID    string
	ReceivedDate  time.Time
	CompletedDate *time.Time
	Image         *string
	Note          *string
	Items         []CreateOrderItemParams
}

type CreateOrderItemParams struct {
	ServiceTypeID string
	Quantity      string
	UnitPrice     string
	Discount      string
	Total         string
}

type CreatePaymentParams struct {
	Amount string
	Method string
	Note   *string
}

type UpsertWorkAssignmentParams struct {
	OrderID     string
	OrderItemID string
	TaskType    string
	EmployeeID  string
	Percent     string
	Amount      string
}

type CreateAttachmentParams struct {
	FilePath  string
	MimeType  *string
	SizeBytes *int
}

type DeliveryRepository interface {
	ListPlans(ctx context.Context, limit int) ([]model.DeliveryPlanListItem, error)
	GetPlan(ctx context.Context, id string) (*model.DeliveryPlanDetail, error)
	CreatePlan(ctx context.Context, p CreatePlanParams) (*model.DeliveryPlanDetail, error)
	DeletePlan(ctx context.Context, id string) error
}

type UserRepository interface {
	List(ctx context.Context) ([]model.User, error)
	Get(ctx context.Context, id string) (*model.User, error)
	GetByEmailForAuth(ctx context.Context, email string) (*UserAuthRow, error)
	Create(ctx context.Context, p CreateUserParams) (*model.User, error)
	Update(ctx context.Context, id string, p UpdateUserParams) (*model.User, error)
	Delete(ctx context.Context, id string) error
}

type UserAuthRow struct {
	User         model.User
	PasswordHash string
}

type CreateUserParams struct {
	Name         string
	Email        string
	Role         string
	PasswordHash string
	IsActive     bool
}

type UpdateUserParams struct {
	Name         string
	Email        string
	Role         string
	PasswordHash *string
	IsActive     bool
}

type CreatePlanParams struct {
	Name         string
	PlannedDate  time.Time
	StartAddress *string
	StartLat     float64
	StartLng     float64
	EndAddress   *string
	EndLat       float64
	EndLng       float64
	Stops        []CreateStopParams
}

type CreateStopParams struct {
	CustomerID string
	Sequence   int
	DistanceKm string
}

type DashboardRepository interface {
	Summary(ctx context.Context, start, end *time.Time) (*model.DashboardSummary, error)
	RevenueSeries(ctx context.Context, start, end time.Time) ([]model.DashboardDailyRow, error)
}

type ReportRepository interface {
	OrdersCSV(ctx context.Context, start, end *time.Time) ([]byte, string, error)
}
