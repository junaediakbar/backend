package model

import "time"

type Customer struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Phone     *string   `json:"phone,omitempty"`
	Address   *string   `json:"address,omitempty"`
	Latitude  *float64  `json:"latitude,omitempty"`
	Longitude *float64  `json:"longitude,omitempty"`
	Email     *string   `json:"email,omitempty"`
	Notes     *string   `json:"notes,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CustomerOrderSummary struct {
	ID             string `json:"id"`
	InvoiceNumber  string `json:"invoiceNumber"`
	Total          string `json:"total"`
	WorkflowStatus string `json:"workflowStatus"`
}

type ServiceType struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Unit         string    `json:"unit"`
	DefaultPrice string    `json:"defaultPrice"`
	IsActive     bool      `json:"isActive"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Employee struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type OrderListItem struct {
	ID            string `json:"id"`
	InvoiceNumber string `json:"invoiceNumber"`
	PublicToken   string `json:"publicToken"`
	Customer      struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"customer"`
	FirstItem *struct {
		ServiceType struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"serviceType"`
	} `json:"firstItem,omitempty"`
	ItemCount      int       `json:"itemCount"`
	Total          string    `json:"total"`
	PaymentStatus  string    `json:"paymentStatus"`
	WorkflowStatus string    `json:"workflowStatus"`
	CreatedAt      time.Time `json:"createdAt"`
}

type OrderDetail struct {
	ID            string `json:"id"`
	InvoiceNumber string `json:"invoiceNumber"`
	PublicToken   string `json:"publicToken"`
	Customer      struct {
		ID    string  `json:"id"`
		Name  string  `json:"name"`
		Phone *string `json:"phone,omitempty"`
	} `json:"customer"`
	Total          string     `json:"total"`
	PaymentStatus  string     `json:"paymentStatus"`
	WorkflowStatus string     `json:"workflowStatus"`
	ReceivedDate   time.Time  `json:"receivedDate"`
	CompletedDate  *time.Time `json:"completedDate,omitempty"`
	PickupDate     *time.Time `json:"pickupDate,omitempty"`
	Image          *string    `json:"image,omitempty"`
	Images         []string   `json:"images,omitempty"`
	Note           *string    `json:"note,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`

	Items       []OrderItem       `json:"items"`
	Payments    []Payment         `json:"payments"`
	Attachments []OrderAttachment `json:"attachments"`
}

type OrderItem struct {
	ID          string `json:"id"`
	ServiceType struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Unit string `json:"unit"`
	} `json:"serviceType"`
	Quantity  string    `json:"quantity"`
	UnitPrice string    `json:"unitPrice"`
	Discount  string    `json:"discount"`
	Total     string    `json:"total"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	WorkAssignments []WorkAssignment `json:"workAssignments"`
}

type WorkAssignment struct {
	ID          string `json:"id"`
	OrderItemID string `json:"orderItemId"`
	TaskType    string `json:"taskType"`
	Employee    struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"employee"`
	Percent   string    `json:"percent"`
	Amount    string    `json:"amount"`
	CreatedAt time.Time `json:"createdAt"`
}

type Payment struct {
	ID        string    `json:"id"`
	OrderID   string    `json:"orderId"`
	Amount    string    `json:"amount"`
	Method    string    `json:"method"`
	PaidAt    time.Time `json:"paidAt"`
	Note      *string   `json:"note,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type OrderAttachment struct {
	ID        string    `json:"id"`
	OrderID   string    `json:"orderId"`
	FilePath  string    `json:"filePath"`
	MimeType  *string   `json:"mimeType,omitempty"`
	SizeBytes *int      `json:"sizeBytes,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type DeliveryPlanListItem struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	PlannedDate time.Time `json:"plannedDate"`
	StopCount   int       `json:"stopCount"`
}

type DeliveryPlanDetail struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	PlannedDate  time.Time `json:"plannedDate"`
	StartAddress *string   `json:"startAddress,omitempty"`
	StartLat     *float64  `json:"startLat,omitempty"`
	StartLng     *float64  `json:"startLng,omitempty"`
	EndAddress   *string   `json:"endAddress,omitempty"`
	EndLat       *float64  `json:"endLat,omitempty"`
	EndLng       *float64  `json:"endLng,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`

	Stops []DeliveryStop `json:"stops"`
}

type DeliveryStop struct {
	ID         string    `json:"id"`
	Sequence   int       `json:"sequence"`
	DistanceKm *string   `json:"distanceKm,omitempty"`
	Customer   Customer  `json:"customer"`
	CreatedAt  time.Time `json:"createdAt"`
}

type DashboardSummary struct {
	CustomerCount int    `json:"customerCount"`
	OrderCount    int    `json:"orderCount"`
	UnpaidCount   int    `json:"unpaidCount"`
	TotalRevenue  string `json:"totalRevenue"`
}

type EmployeePerformanceRow struct {
	EmployeeID   string `json:"employeeId"`
	EmployeeName string `json:"employeeName"`
	PickupAmount string `json:"pickupAmount"`
	WorkAmount   string `json:"workAmount"`
	TotalAmount  string `json:"totalAmount"`
}

type Paged[T any] struct {
	Items          []T    `json:"items"`
	Page           int    `json:"page"`
	PageSize       int    `json:"pageSize"`
	Total          int    `json:"total"`
	RevenueTotal   string `json:"revenueTotal,omitempty"` // SUM(order.total) for current filter (numeric string)
}
