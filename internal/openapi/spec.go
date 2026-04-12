package openapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"

	"laundry-backend/internal/httpapi"
	"laundry-backend/internal/model"
)

func JSON() ([]byte, error) {
	spec := &openapi3.T{
		OpenAPI: "3.0.3",
		Info: &openapi3.Info{
			Title:   "Laundry API",
			Version: "1.0.0",
		},
		Paths: openapi3.NewPaths(),
		Components: &openapi3.Components{
			SecuritySchemes: openapi3.SecuritySchemes{
				"bearerAuth": &openapi3.SecuritySchemeRef{
					Value: &openapi3.SecurityScheme{
						Type:         "http",
						Scheme:       "bearer",
						BearerFormat: "JWT",
					},
				},
				"apiKeyAuth": &openapi3.SecuritySchemeRef{
					Value: &openapi3.SecurityScheme{
						Type: "apiKey",
						In:   "header",
						Name: "X-API-Key",
					},
				},
			},
			Schemas: openapi3.Schemas{},
		},
	}

	type okResp struct {
		OK bool `json:"ok"`
	}

	type healthData struct {
		Status string `json:"status"`
	}

	type authLoginBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type authLoginData struct {
		Token string     `json:"token"`
		User  model.User `json:"user"`
	}

	type authMeData struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Role  string `json:"role"`
	}

	type customerUpsertBody struct {
		Name      string   `json:"name"`
		Phone     *string  `json:"phone"`
		Address   *string  `json:"address"`
		Latitude  *float64 `json:"latitude"`
		Longitude *float64 `json:"longitude"`
		Email     *string  `json:"email"`
		Notes     *string  `json:"notes"`
	}

	type employeeUpsertBody struct {
		Name     string `json:"name"`
		IsActive bool   `json:"isActive"`
	}

	type serviceTypeUpsertBody struct {
		Name         string  `json:"name"`
		Unit         string  `json:"unit"`
		DefaultPrice float64 `json:"defaultPrice"`
		IsActive     bool    `json:"isActive"`
	}

	type userUpsertBody struct {
		Name     string  `json:"name"`
		Email    string  `json:"email"`
		Role     string  `json:"role"`
		Password *string `json:"password"`
		IsActive bool    `json:"isActive"`
	}

	type workflowBody struct {
		WorkflowStatus string `json:"workflowStatus"`
	}

	type paymentBody struct {
		Amount float64 `json:"amount"`
		Method string  `json:"method"`
		Note   *string `json:"note"`
	}

	type attachmentsBody struct {
		Files []struct {
			FilePath  string  `json:"filePath"`
			MimeType  *string `json:"mimeType"`
			SizeBytes *int    `json:"sizeBytes"`
		} `json:"files"`
	}

	type workAssignmentBody struct {
		EmployeeID string   `json:"employeeId"`
		Percent    *float64 `json:"percent"`
	}

	type deliveryPlanCreateBody struct {
		Name         string  `json:"name"`
		PlannedDate  string  `json:"plannedDate"`
		StartAddress *string `json:"startAddress"`
		StartLat     float64 `json:"startLat"`
		StartLng     float64 `json:"startLng"`
		EndAddress   *string `json:"endAddress"`
		EndLat       float64 `json:"endLat"`
		EndLng       float64 `json:"endLng"`
		Stops        []struct {
			CustomerID string  `json:"customerId"`
			Sequence   int     `json:"sequence"`
			DistanceKm float64 `json:"distanceKm"`
		} `json:"stops"`
	}

	type publicReceiptItem struct {
		ServiceName string  `json:"serviceName"`
		Unit        string  `json:"unit"`
		Quantity    string  `json:"quantity"`
		UnitPrice   string  `json:"unitPrice"`
		Discount    string  `json:"discount"`
		Total       string  `json:"total"`
		LengthM     *string `json:"lengthM,omitempty"`
		WidthM      *string `json:"widthM,omitempty"`
	}

	type publicReceipt struct {
		PublicToken    string             `json:"publicToken"`
		InvoiceNumber  string             `json:"invoiceNumber"`
		CustomerName   string             `json:"customerName"`
		CustomerPhone  *string            `json:"customerPhone,omitempty"`
		Total          string             `json:"total"`
		PaidAmount     string             `json:"paidAmount"`
		PaymentStatus  string             `json:"paymentStatus"`
		WorkflowStatus string             `json:"workflowStatus"`
		ReceivedDate   string             `json:"receivedDate"`
		CompletedDate  *string            `json:"completedDate,omitempty"`
		PickupDate     *string            `json:"pickupDate,omitempty"`
		Image          *string            `json:"image,omitempty"`
		Images         []string           `json:"images,omitempty"`
		Note           *string            `json:"note,omitempty"`
		Items          []publicReceiptItem `json:"items"`
	}

	var firstErr error
	gen := openapi3gen.NewGenerator(
		openapi3gen.CreateComponentSchemas(openapi3gen.ExportComponentSchemasOptions{
			ExportComponentSchemas: true,
			ExportTopLevelSchema:   true,
			ExportGenerics:         true,
		}),
	)

	schemaFor := func(v any) *openapi3.SchemaRef {
		ref, err := gen.NewSchemaRefForValue(v, spec.Components.Schemas)
		if err != nil && firstErr == nil {
			firstErr = err
		}
		if ref == nil {
			return &openapi3.SchemaRef{Value: openapi3.NewSchema()}
		}
		return ref
	}

	errorSchema := schemaFor(httpapi.Err{})
	errorEnvelopeSchema := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type:     &openapi3.Types{"object"},
			Required: []string{"ok", "error"},
			Properties: openapi3.Schemas{
				"ok": {
					Value: &openapi3.Schema{
						Type: &openapi3.Types{"boolean"},
						Enum: []any{false},
					},
				},
				"error": errorSchema,
			},
		},
	}
	spec.Components.Schemas["ApiErrorEnvelope"] = errorEnvelopeSchema

	okEnvelope := func(data *openapi3.SchemaRef) *openapi3.SchemaRef {
		return &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type:     &openapi3.Types{"object"},
				Required: []string{"ok", "data"},
				Properties: openapi3.Schemas{
					"ok": {
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"boolean"},
							Enum: []any{true},
						},
					},
					"data": data,
				},
			},
		}
	}

	jsonResponse := func(desc string, schema *openapi3.SchemaRef) *openapi3.ResponseRef {
		return &openapi3.ResponseRef{
			Value: &openapi3.Response{
				Description: &desc,
				Content: openapi3.Content{
					"application/json": &openapi3.MediaType{Schema: schema},
				},
			},
		}
	}

	addStandardErrors := func(op *openapi3.Operation) {
		desc := "Error"
		op.Responses.Set(strconv.Itoa(http.StatusBadRequest), jsonResponse(desc, &openapi3.SchemaRef{Ref: "#/components/schemas/ApiErrorEnvelope"}))
		op.Responses.Set(strconv.Itoa(http.StatusUnauthorized), jsonResponse(desc, &openapi3.SchemaRef{Ref: "#/components/schemas/ApiErrorEnvelope"}))
		op.Responses.Set(strconv.Itoa(http.StatusForbidden), jsonResponse(desc, &openapi3.SchemaRef{Ref: "#/components/schemas/ApiErrorEnvelope"}))
		op.Responses.Set(strconv.Itoa(http.StatusNotFound), jsonResponse(desc, &openapi3.SchemaRef{Ref: "#/components/schemas/ApiErrorEnvelope"}))
		op.Responses.Set(strconv.Itoa(http.StatusConflict), jsonResponse(desc, &openapi3.SchemaRef{Ref: "#/components/schemas/ApiErrorEnvelope"}))
		op.Responses.Set(strconv.Itoa(http.StatusInternalServerError), jsonResponse(desc, &openapi3.SchemaRef{Ref: "#/components/schemas/ApiErrorEnvelope"}))
	}

	queryParam := func(name string, schema *openapi3.SchemaRef, desc string) *openapi3.ParameterRef {
		return &openapi3.ParameterRef{
			Value: &openapi3.Parameter{
				In:          "query",
				Name:        name,
				Description: desc,
				Schema:      schema,
			},
		}
	}

	pathParam := func(name string, schema *openapi3.SchemaRef, desc string) *openapi3.ParameterRef {
		return &openapi3.ParameterRef{
			Value: &openapi3.Parameter{
				In:          "path",
				Name:        name,
				Required:    true,
				Description: desc,
				Schema:      schema,
			},
		}
	}

	addOp := func(method, path, summary string, secured bool) *openapi3.Operation {
		op := &openapi3.Operation{Summary: summary, Responses: openapi3.NewResponses()}
		if secured {
			op.Security = &openapi3.SecurityRequirements{
				{"bearerAuth": {}},
				{"apiKeyAuth": {}},
			}
		}
		pi := spec.Paths.Value(path)
		if pi == nil {
			pi = &openapi3.PathItem{}
			spec.Paths.Set(path, pi)
		}
		switch method {
		case http.MethodGet:
			pi.Get = op
		case http.MethodPost:
			pi.Post = op
		case http.MethodPut:
			pi.Put = op
		case http.MethodPatch:
			pi.Patch = op
		case http.MethodDelete:
			pi.Delete = op
		default:
			if firstErr == nil {
				firstErr = fmt.Errorf("unsupported method: %s", method)
			}
		}
		return op
	}

	addJSON := func(op *openapi3.Operation, status int, v any) {
		desc := "OK"
		op.Responses.Set(strconv.Itoa(status), jsonResponse(desc, okEnvelope(schemaFor(v))))
		addStandardErrors(op)
	}

	addBodyJSON := func(op *openapi3.Operation, v any) {
		op.RequestBody = &openapi3.RequestBodyRef{
			Value: &openapi3.RequestBody{
				Required: true,
				Content: openapi3.Content{
					"application/json": &openapi3.MediaType{Schema: schemaFor(v)},
				},
			},
		}
	}

	addBodyMultipartOrder := func(op *openapi3.Operation) {
		s := &openapi3.Schema{
			Type: &openapi3.Types{"object"},
			Properties: openapi3.Schemas{
				"customerId":    {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
				"receivedDate":  {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "date-time"}},
				"completedDate": {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "date-time"}},
				"note":          {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
				"items":         {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Description: "JSON string array of items"}},
				"images": {
					Value: &openapi3.Schema{
						Type: &openapi3.Types{"array"},
						Items: &openapi3.SchemaRef{
							Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "binary"},
						},
					},
				},
			},
			Required: []string{"customerId", "items"},
		}
		op.RequestBody = &openapi3.RequestBodyRef{
			Value: &openapi3.RequestBody{
				Required: true,
				Content: openapi3.Content{
					"multipart/form-data": &openapi3.MediaType{Schema: &openapi3.SchemaRef{Value: s}},
				},
			},
		}
	}

	addJSON(addOp(http.MethodGet, "/health", "Health check", false), http.StatusOK, healthData{})
	{
		desc := "OpenAPI spec"
		addOp(http.MethodGet, "/openapi.json", "OpenAPI spec", false).Responses.Set(
			strconv.Itoa(http.StatusOK),
			&openapi3.ResponseRef{Value: &openapi3.Response{Description: &desc}},
		)
	}

	addJSON(addOp(http.MethodGet, "/api/v1/auth/me", "Get current user session", true), http.StatusOK, authMeData{})
	opLogin := addOp(http.MethodPost, "/api/v1/auth/login", "Login", false)
	addBodyJSON(opLogin, authLoginBody{})
	addJSON(opLogin, http.StatusOK, authLoginData{})

	opPubReceipt := addOp(http.MethodGet, "/api/v1/public/receipts/{token}", "Get public receipt by token", false)
	opPubReceipt.Parameters = append(opPubReceipt.Parameters, pathParam("token", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}, "Public token"))
	addJSON(opPubReceipt, http.StatusOK, publicReceipt{})

	addJSON(addOp(http.MethodGet, "/api/v1/dashboard/summary", "Dashboard summary", true), http.StatusOK, model.DashboardSummary{})

	opCustomersList := addOp(http.MethodGet, "/api/v1/customers", "List customers", true)
	opCustomersList.Parameters = append(opCustomersList.Parameters,
		queryParam("q", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}, "Search query"),
		queryParam("page", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}}}, "Page number"),
		queryParam("pageSize", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}}}, "Items per page"),
	)
	addJSON(opCustomersList, http.StatusOK, model.Paged[model.Customer]{})

	opCustomersCreate := addOp(http.MethodPost, "/api/v1/customers", "Create customer", true)
	addBodyJSON(opCustomersCreate, customerUpsertBody{})
	addJSON(opCustomersCreate, http.StatusCreated, model.Customer{})

	opCustomerGet := addOp(http.MethodGet, "/api/v1/customers/{id}", "Get customer", true)
	opCustomerGet.Parameters = append(opCustomerGet.Parameters, pathParam("id", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}, "Customer ID"))
	addJSON(opCustomerGet, http.StatusOK, model.Customer{})

	opCustomerUpdate := addOp(http.MethodPut, "/api/v1/customers/{id}", "Update customer", true)
	opCustomerUpdate.Parameters = append(opCustomerUpdate.Parameters, pathParam("id", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}, "Customer ID"))
	addBodyJSON(opCustomerUpdate, customerUpsertBody{})
	addJSON(opCustomerUpdate, http.StatusOK, model.Customer{})

	opCustomerDelete := addOp(http.MethodDelete, "/api/v1/customers/{id}", "Delete customer", true)
	opCustomerDelete.Parameters = append(opCustomerDelete.Parameters, pathParam("id", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}, "Customer ID"))
	addJSON(opCustomerDelete, http.StatusOK, okResp{OK: true})

	opCustomerRecent := addOp(http.MethodGet, "/api/v1/customers/{id}/orders", "List recent customer orders", true)
	opCustomerRecent.Parameters = append(opCustomerRecent.Parameters,
		pathParam("id", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}, "Customer ID"),
		queryParam("limit", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}}}, "Max items"),
	)
	addJSON(opCustomerRecent, http.StatusOK, []model.CustomerOrderSummary{})

	opServiceTypesList := addOp(http.MethodGet, "/api/v1/service-types", "List service types", true)
	opServiceTypesList.Parameters = append(opServiceTypesList.Parameters, queryParam("active", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"boolean"}}}, "Only active service types"))
	addJSON(opServiceTypesList, http.StatusOK, []model.ServiceType{})

	opServiceTypesCreate := addOp(http.MethodPost, "/api/v1/service-types", "Create service type", true)
	addBodyJSON(opServiceTypesCreate, serviceTypeUpsertBody{})
	addJSON(opServiceTypesCreate, http.StatusCreated, model.ServiceType{})

	opServiceTypesGet := addOp(http.MethodGet, "/api/v1/service-types/{id}", "Get service type", true)
	opServiceTypesGet.Parameters = append(opServiceTypesGet.Parameters, pathParam("id", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}, "Service type ID"))
	addJSON(opServiceTypesGet, http.StatusOK, model.ServiceType{})

	opServiceTypesUpdate := addOp(http.MethodPut, "/api/v1/service-types/{id}", "Update service type", true)
	opServiceTypesUpdate.Parameters = append(opServiceTypesUpdate.Parameters, pathParam("id", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}, "Service type ID"))
	addBodyJSON(opServiceTypesUpdate, serviceTypeUpsertBody{})
	addJSON(opServiceTypesUpdate, http.StatusOK, model.ServiceType{})

	opServiceTypesDelete := addOp(http.MethodDelete, "/api/v1/service-types/{id}", "Delete service type", true)
	opServiceTypesDelete.Parameters = append(opServiceTypesDelete.Parameters, pathParam("id", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}, "Service type ID"))
	addJSON(opServiceTypesDelete, http.StatusOK, okResp{OK: true})

	opEmployeesList := addOp(http.MethodGet, "/api/v1/employees", "List employees", true)
	opEmployeesList.Parameters = append(opEmployeesList.Parameters, queryParam("active", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"boolean"}}}, "Only active employees"))
	addJSON(opEmployeesList, http.StatusOK, []model.Employee{})

	opEmployeesPerf := addOp(http.MethodGet, "/api/v1/employees/performance", "Employee performance", true)
	opEmployeesPerf.Parameters = append(opEmployeesPerf.Parameters,
		queryParam("startDate", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "date"}}, "YYYY-MM-DD"),
		queryParam("endDate", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "date"}}, "YYYY-MM-DD"),
	)
	addJSON(opEmployeesPerf, http.StatusOK, []model.EmployeePerformanceRow{})

	opEmployeesCreate := addOp(http.MethodPost, "/api/v1/employees", "Create employee", true)
	addBodyJSON(opEmployeesCreate, employeeUpsertBody{})
	addJSON(opEmployeesCreate, http.StatusCreated, model.Employee{})

	opEmployeesGet := addOp(http.MethodGet, "/api/v1/employees/{id}", "Get employee", true)
	opEmployeesGet.Parameters = append(opEmployeesGet.Parameters, pathParam("id", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}, "Employee ID"))
	addJSON(opEmployeesGet, http.StatusOK, model.Employee{})

	opEmployeesUpdate := addOp(http.MethodPut, "/api/v1/employees/{id}", "Update employee", true)
	opEmployeesUpdate.Parameters = append(opEmployeesUpdate.Parameters, pathParam("id", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}, "Employee ID"))
	addBodyJSON(opEmployeesUpdate, employeeUpsertBody{})
	addJSON(opEmployeesUpdate, http.StatusOK, model.Employee{})

	opEmployeesDelete := addOp(http.MethodDelete, "/api/v1/employees/{id}", "Delete employee", true)
	opEmployeesDelete.Parameters = append(opEmployeesDelete.Parameters, pathParam("id", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}, "Employee ID"))
	addJSON(opEmployeesDelete, http.StatusOK, okResp{OK: true})

	opOrdersList := addOp(http.MethodGet, "/api/v1/orders", "List orders", true)
	opOrdersList.Parameters = append(opOrdersList.Parameters,
		queryParam("q", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}, "Search query"),
		queryParam("page", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}}}, "Page number"),
		queryParam("pageSize", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}}}, "Items per page"),
		queryParam("sort", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}, "Sort field"),
		queryParam("dir", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}, "Sort direction"),
		queryParam("startDate", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "date"}}, "YYYY-MM-DD"),
		queryParam("endDate", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "date"}}, "YYYY-MM-DD"),
	)
	addJSON(opOrdersList, http.StatusOK, model.Paged[model.OrderListItem]{})

	opOrdersCreate := addOp(http.MethodPost, "/api/v1/orders", "Create order", true)
	addBodyMultipartOrder(opOrdersCreate)
	addJSON(opOrdersCreate, http.StatusCreated, model.OrderDetail{})

	opOrderGet := addOp(http.MethodGet, "/api/v1/orders/{id}", "Get order detail", true)
	opOrderGet.Parameters = append(opOrderGet.Parameters, pathParam("id", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}, "Order ID"))
	addJSON(opOrderGet, http.StatusOK, model.OrderDetail{})

	opOrderDelete := addOp(http.MethodDelete, "/api/v1/orders/{id}", "Delete order", true)
	opOrderDelete.Parameters = append(opOrderDelete.Parameters, pathParam("id", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}, "Order ID"))
	addJSON(opOrderDelete, http.StatusOK, okResp{OK: true})

	opOrderWorkflow := addOp(http.MethodPatch, "/api/v1/orders/{id}/workflow", "Update workflow", true)
	opOrderWorkflow.Parameters = append(opOrderWorkflow.Parameters, pathParam("id", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}, "Order ID"))
	addBodyJSON(opOrderWorkflow, workflowBody{})
	addJSON(opOrderWorkflow, http.StatusOK, okResp{OK: true})

	opOrderPayment := addOp(http.MethodPost, "/api/v1/orders/{id}/payments", "Create payment", true)
	opOrderPayment.Parameters = append(opOrderPayment.Parameters, pathParam("id", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}, "Order ID"))
	addBodyJSON(opOrderPayment, paymentBody{})
	addJSON(opOrderPayment, http.StatusCreated, model.Payment{})

	opOrderPaymentDelete := addOp(http.MethodDelete, "/api/v1/orders/{id}/payments/{paymentId}", "Delete payment", true)
	opOrderPaymentDelete.Parameters = append(opOrderPaymentDelete.Parameters,
		pathParam("id", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}, "Order ID"),
		pathParam("paymentId", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}, "Payment ID"),
	)
	addJSON(opOrderPaymentDelete, http.StatusOK, model.Payment{})

	opOrderAttachments := addOp(http.MethodPost, "/api/v1/orders/{id}/attachments", "Create attachments", true)
	opOrderAttachments.Parameters = append(opOrderAttachments.Parameters, pathParam("id", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}, "Order ID"))
	addBodyJSON(opOrderAttachments, attachmentsBody{})
	addJSON(opOrderAttachments, http.StatusOK, okResp{OK: true})

	opUpsertWork := addOp(http.MethodPut, "/api/v1/orders/{orderId}/items/{orderItemId}/work-assignments/{taskType}", "Upsert work assignment", true)
	opUpsertWork.Parameters = append(opUpsertWork.Parameters,
		pathParam("orderId", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}, "Order ID"),
		pathParam("orderItemId", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}, "Order item ID"),
		pathParam("taskType", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}, "Task type"),
	)
	addBodyJSON(opUpsertWork, workAssignmentBody{})
	addJSON(opUpsertWork, http.StatusOK, okResp{OK: true})

	opPlansList := addOp(http.MethodGet, "/api/v1/delivery-plans", "List delivery plans", true)
	opPlansList.Parameters = append(opPlansList.Parameters, queryParam("limit", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}}}, "Max items"))
	addJSON(opPlansList, http.StatusOK, []model.DeliveryPlanListItem{})

	opPlansCreate := addOp(http.MethodPost, "/api/v1/delivery-plans", "Create delivery plan", true)
	addBodyJSON(opPlansCreate, deliveryPlanCreateBody{})
	addJSON(opPlansCreate, http.StatusCreated, model.DeliveryPlanDetail{})

	opPlansGet := addOp(http.MethodGet, "/api/v1/delivery-plans/{id}", "Get delivery plan", true)
	opPlansGet.Parameters = append(opPlansGet.Parameters, pathParam("id", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}, "Delivery plan ID"))
	addJSON(opPlansGet, http.StatusOK, model.DeliveryPlanDetail{})

	opPlansDelete := addOp(http.MethodDelete, "/api/v1/delivery-plans/{id}", "Delete delivery plan", true)
	opPlansDelete.Parameters = append(opPlansDelete.Parameters, pathParam("id", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}, "Delivery plan ID"))
	addJSON(opPlansDelete, http.StatusOK, okResp{OK: true})

	opReportCSV := addOp(http.MethodGet, "/api/v1/reports/orders.csv", "Export orders CSV", true)
	opReportCSV.Parameters = append(opReportCSV.Parameters,
		queryParam("startDate", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "date"}}, "YYYY-MM-DD"),
		queryParam("endDate", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "date"}}, "YYYY-MM-DD"),
	)
	descCSV := "CSV file"
	opReportCSV.Responses.Set(strconv.Itoa(http.StatusOK), &openapi3.ResponseRef{
		Value: &openapi3.Response{
			Description: &descCSV,
			Content: openapi3.Content{
				"text/csv": &openapi3.MediaType{
					Schema: &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "binary"}},
				},
			},
		},
	})
	addStandardErrors(opReportCSV)

	opUsersList := addOp(http.MethodGet, "/api/v1/users", "List users", true)
	addJSON(opUsersList, http.StatusOK, []model.User{})

	opUsersCreate := addOp(http.MethodPost, "/api/v1/users", "Create user", true)
	addBodyJSON(opUsersCreate, userUpsertBody{})
	addJSON(opUsersCreate, http.StatusCreated, model.User{})

	opUsersGet := addOp(http.MethodGet, "/api/v1/users/{id}", "Get user", true)
	opUsersGet.Parameters = append(opUsersGet.Parameters, pathParam("id", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}, "User ID"))
	addJSON(opUsersGet, http.StatusOK, model.User{})

	opUsersUpdate := addOp(http.MethodPut, "/api/v1/users/{id}", "Update user", true)
	opUsersUpdate.Parameters = append(opUsersUpdate.Parameters, pathParam("id", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}, "User ID"))
	addBodyJSON(opUsersUpdate, userUpsertBody{})
	addJSON(opUsersUpdate, http.StatusOK, model.User{})

	opUsersDelete := addOp(http.MethodDelete, "/api/v1/users/{id}", "Delete user", true)
	opUsersDelete.Parameters = append(opUsersDelete.Parameters, pathParam("id", &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}, "User ID"))
	addJSON(opUsersDelete, http.StatusOK, okResp{OK: true})

	if firstErr != nil {
		return nil, firstErr
	}

	return json.MarshalIndent(spec, "", "  ")
}
