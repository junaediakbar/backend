package openapi

import (
	"encoding/json"

	"github.com/getkin/kin-openapi/openapi3"
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
		},
	}

	add := func(method, path, summary string, secured bool) {
		op := &openapi3.Operation{Summary: summary}
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
		case "GET":
			pi.Get = op
		case "POST":
			pi.Post = op
		case "PUT":
			pi.Put = op
		case "PATCH":
			pi.Patch = op
		case "DELETE":
			pi.Delete = op
		}
	}

	add("GET", "/health", "Health check", false)
	add("GET", "/openapi.json", "OpenAPI spec", false)

	add("GET", "/api/v1/dashboard/summary", "Dashboard summary", true)

	add("GET", "/api/v1/customers", "List customers", true)
	add("POST", "/api/v1/customers", "Create customer", true)
	add("GET", "/api/v1/customers/{id}", "Get customer", true)
	add("PUT", "/api/v1/customers/{id}", "Update customer", true)
	add("DELETE", "/api/v1/customers/{id}", "Delete customer", true)
	add("GET", "/api/v1/customers/{id}/orders", "List recent customer orders", true)

	add("GET", "/api/v1/service-types", "List service types", true)
	add("POST", "/api/v1/service-types", "Create service type", true)
	add("GET", "/api/v1/service-types/{id}", "Get service type", true)
	add("PUT", "/api/v1/service-types/{id}", "Update service type", true)
	add("DELETE", "/api/v1/service-types/{id}", "Delete service type", true)

	add("GET", "/api/v1/employees", "List employees", true)
	add("GET", "/api/v1/employees/performance", "Employee performance", true)
	add("POST", "/api/v1/employees", "Create employee", true)
	add("GET", "/api/v1/employees/{id}", "Get employee", true)
	add("PUT", "/api/v1/employees/{id}", "Update employee", true)
	add("DELETE", "/api/v1/employees/{id}", "Delete employee", true)

	add("GET", "/api/v1/orders", "List orders", true)
	add("POST", "/api/v1/orders", "Create order", true)
	add("GET", "/api/v1/orders/{id}", "Get order detail", true)
	add("DELETE", "/api/v1/orders/{id}", "Delete order", true)
	add("PATCH", "/api/v1/orders/{id}/workflow", "Update workflow", true)
	add("POST", "/api/v1/orders/{id}/payments", "Create payment", true)
	add("POST", "/api/v1/orders/{id}/attachments", "Create attachments", true)
	add("PUT", "/api/v1/orders/{orderId}/items/{orderItemId}/work-assignments/{taskType}", "Upsert work assignment", true)

	add("GET", "/api/v1/delivery-plans", "List delivery plans", true)
	add("POST", "/api/v1/delivery-plans", "Create delivery plan", true)
	add("GET", "/api/v1/delivery-plans/{id}", "Get delivery plan", true)
	add("DELETE", "/api/v1/delivery-plans/{id}", "Delete delivery plan", true)

	add("GET", "/api/v1/reports/orders.csv", "Export orders CSV", true)
	add("GET", "/api/v1/users", "List users", true)
	add("POST", "/api/v1/users", "Create user", true)
	add("GET", "/api/v1/users/{id}", "Get user", true)
	add("PUT", "/api/v1/users/{id}", "Update user", true)
	add("DELETE", "/api/v1/users/{id}", "Delete user", true)

	return json.MarshalIndent(spec, "", "  ")
}
