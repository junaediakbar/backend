package httpserver

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/swaggest/swgui/v5emb"

	"laundry-backend/internal/http/handler"
	"laundry-backend/internal/http/middleware"
)

type ServerDeps struct {
	Auth middleware.AuthConfig

	Authn          *handler.AuthHandler
	PublicReceipts *handler.PublicReceiptHandler
	Dashboard      *handler.DashboardHandler
	Customers      *handler.CustomerHandler
	Orders         *handler.OrderHandler
	ServiceTypes   *handler.ServiceTypeHandler
	Employees      *handler.EmployeeHandler
	Delivery       *handler.DeliveryHandler
	Reports        *handler.ReportHandler
}

func NewRouter(deps ServerDeps) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.RealIP)
	r.Use(chimw.RequestID)
	r.Use(middleware.RequestLogger())
	r.Use(middleware.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))

	r.Get("/health", handler.Health().ServeHTTP)
	r.Get("/openapi.json", handler.OpenAPIJSON().ServeHTTP)
	r.Mount("/docs", v5emb.New("Laundry API", "/openapi.json", "/docs/"))

	jwksProvider := &middleware.JWKSProvider{}

	r.Route("/api/v1", func(api chi.Router) {
		if deps.PublicReceipts != nil {
			api.Route("/public", func(pr chi.Router) {
				pr.Get("/receipts/{token}", deps.PublicReceipts.Get().ServeHTTP)
			})
		}

		if deps.Authn != nil {
			api.Route("/auth", func(ar chi.Router) {
				ar.Post("/login", deps.Authn.Login().ServeHTTP)

				ar.Group(func(pr chi.Router) {
					pr.Use(middleware.WithAuth(deps.Auth, jwksProvider))
					pr.Get("/me", deps.Authn.Me().ServeHTTP)
				})
			})
		}

		api.Group(func(pr chi.Router) {
			pr.Use(middleware.WithAuth(deps.Auth, jwksProvider))

			// Karyawan: hanya GET pada /orders (lihat nota).
			pr.Route("/orders", func(or chi.Router) {
				or.Use(middleware.OrderEmployeeReadOnly)
				or.Get("/", deps.Orders.List().ServeHTTP)
				or.Post("/", deps.Orders.Create().ServeHTTP)
				or.Route("/{id}", func(ir chi.Router) {
					ir.Get("/", deps.Orders.Get().ServeHTTP)
					ir.Delete("/", deps.Orders.Delete().ServeHTTP)
					ir.Patch("/workflow", deps.Orders.UpdateWorkflow().ServeHTTP)
					ir.Post("/payments", deps.Orders.CreatePayment().ServeHTTP)
					ir.Delete("/payments/{paymentId}", deps.Orders.DeletePayment().ServeHTTP)
					ir.Post("/attachments", deps.Orders.CreateAttachments().ServeHTTP)
				})
				or.Put("/{orderId}/items/{orderItemId}/work-assignments/{taskType}", deps.Orders.UpsertWorkAssignment().ServeHTTP)
			})

			// Kinerja: karyawan hanya melihat data sendiri (filter di handler).
			pr.Route("/employees", func(er chi.Router) {
				er.Get("/performance", deps.Employees.Performance().ServeHTTP)
				// Daftar nama untuk penugasan nota — semua role yang login; karyawan dapat versi tanpa email/role.
				er.Get("/", deps.Employees.List().ServeHTTP)
				er.Group(func(ier chi.Router) {
					ier.Use(middleware.RequireNotEmployee)
					ier.Post("/", deps.Employees.Create().ServeHTTP)
					ier.Route("/{id}", func(ir chi.Router) {
						ir.Get("/", deps.Employees.Get().ServeHTTP)
						ir.Put("/", deps.Employees.Update().ServeHTTP)
						ir.Delete("/", deps.Employees.Delete().ServeHTTP)
					})
				})
			})

			pr.Group(func(staff chi.Router) {
				staff.Use(middleware.RequireNotEmployee)

				staff.Get("/dashboard/summary", deps.Dashboard.Summary().ServeHTTP)
				staff.Get("/dashboard/revenue-series", deps.Dashboard.RevenueSeries().ServeHTTP)

				staff.Route("/customers", func(cr chi.Router) {
					cr.Get("/", deps.Customers.List().ServeHTTP)
					cr.Post("/", deps.Customers.Create().ServeHTTP)
					cr.Delete("/{id}", deps.Customers.Delete().ServeHTTP)
					cr.Delete("/{id}/", deps.Customers.Delete().ServeHTTP)
					cr.Route("/{id}", func(ir chi.Router) {
						ir.Get("/", deps.Customers.Get().ServeHTTP)
						ir.Get("/orders", deps.Customers.RecentOrders().ServeHTTP)
						ir.Put("/", deps.Customers.Update().ServeHTTP)
						ir.Delete("/", deps.Customers.Delete().ServeHTTP)
					})
				})

				staff.Route("/service-types", func(sr chi.Router) {
					sr.Get("/", deps.ServiceTypes.List().ServeHTTP)
					sr.Post("/", deps.ServiceTypes.Create().ServeHTTP)
					sr.Route("/{id}", func(ir chi.Router) {
						ir.Get("/", deps.ServiceTypes.Get().ServeHTTP)
						ir.Put("/", deps.ServiceTypes.Update().ServeHTTP)
						ir.Delete("/", deps.ServiceTypes.Delete().ServeHTTP)
					})
				})

				staff.Route("/delivery-plans", func(dr chi.Router) {
					dr.Get("/", deps.Delivery.ListPlans().ServeHTTP)
					dr.Post("/", deps.Delivery.CreatePlan().ServeHTTP)
					dr.Get("/{id}", deps.Delivery.GetPlan().ServeHTTP)
					dr.Delete("/{id}", deps.Delivery.DeletePlan().ServeHTTP)
					dr.Delete("/{id}/", deps.Delivery.DeletePlan().ServeHTTP)
				})

				staff.Get("/reports/orders.csv", deps.Reports.OrdersCSV().ServeHTTP)
			})
		})
	})

	return r
}
