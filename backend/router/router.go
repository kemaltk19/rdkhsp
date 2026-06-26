package router

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"radikal-hesap/database"
	"radikal-hesap/handlers"
	"radikal-hesap/middleware"
	"radikal-hesap/services"
	"radikal-hesap/utils"
)

// Setup configures CORS, public/private groups, and maps endpoint handlers.
func Setup() *gin.Engine {
	r := gin.Default()

	// CORS Middleware
	r.Use(CORS())

	health := handlers.NewHealthHandler(database.DB)
	authSvc := services.NewAuthService()
	auth := handlers.NewAuthHandler(authSvc)
	cariSvc := services.NewCariService()
	cari := handlers.NewCariHandler(cariSvc)
	invoiceSvc := services.NewInvoiceService()
	invoice := handlers.NewInvoiceHandler(invoiceSvc)
	paymentSvc := services.NewPaymentService()
	cashAccSvc := services.NewCashAccountService()
	bankAccSvc := services.NewBankAccountService()
	payment := handlers.NewPaymentHandler(paymentSvc, cashAccSvc, bankAccSvc)
	expenseSvc := services.NewExpenseService()
	expenseCatSvc := services.NewExpenseCategoryService()
	expense := handlers.NewExpenseHandler(expenseSvc, expenseCatSvc)
	stockSvc := services.NewStockService()
	product := handlers.NewProductHandler(stockSvc)
	cashSvc := services.NewCashService()
	cash := handlers.NewCashHandler(cashSvc)
	quoteSvc := services.NewQuoteService()
	quote := handlers.NewQuoteHandler(quoteSvc)
	dashboardSvc := services.NewDashboardService()
	dashboard := handlers.NewDashboardHandler(dashboardSvc)
	reportSvc := services.NewReportService()
	report := handlers.NewReportHandler(reportSvc)
	employeeSvc := services.NewEmployeeService()
	employee := handlers.NewEmployeeHandler(employeeSvc)
	settingsSvc := services.NewSettingsService()
	settings := handlers.NewSettingsHandler(settingsSvc)
	currencySvc := services.NewCurrencyService()
	currency := handlers.NewCurrencyHandler(currencySvc)
	billingSvc := services.NewBillingService(&services.MockPaymentProvider{})
	billing := handlers.NewBillingHandler(billingSvc)
	importSvc := services.NewImportService(database.DB)
	importHndlr := handlers.NewImportHandler(importSvc)
	roleSvc := services.NewRoleService()
	role := handlers.NewRoleHandler(roleSvc)
	audit := handlers.NewAuditHandler()
	publicInvoice := handlers.NewPublicInvoiceHandler()
	publicQuote := handlers.NewPublicQuoteHandler(quoteSvc)
	notifSvc := services.NewNotificationService()
	notification := handlers.NewNotificationHandler(notifSvc)
	projectSvc := services.NewProjectService()
	project := handlers.NewProjectHandler(projectSvc)
	projectCategorySvc := services.NewProjectCategoryService()
	projectCategory := handlers.NewProjectCategoryHandler(projectCategorySvc)

	api := r.Group("/api")
	{
		// Health check
		api.GET("/health", health.Health)
		api.GET("/health/db", health.HealthDB)

		// Public Auth routes
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", auth.Register)
			authGroup.POST("/login", auth.Login)
			authGroup.POST("/refresh", auth.Refresh)
			authGroup.POST("/logout", auth.Logout)
		}

		// Public Billing Webhook
		api.POST("/billing/webhook", billing.ProcessWebhook)

		// Public Invoices
		api.GET("/public/invoices/:token", publicInvoice.GetByToken)
		api.POST("/public/invoices/:token/dispute", publicInvoice.Dispute)
		api.POST("/public/invoices/:token/pay", publicInvoice.Pay)

		// Public Quotes
		api.GET("/public/quotes/:token", publicQuote.GetByToken)
		api.POST("/public/quotes/:token/accept", publicQuote.AcceptPublic)
		api.POST("/public/quotes/:token/reject", publicQuote.RejectPublic)

		// Auth + Tenant (Subscription optional - e.g. for /auth/me)
		authTenant := api.Group("")
		authTenant.Use(middleware.AuthMiddleware(), middleware.TenantMiddleware())
		{
			authTenant.GET("/auth/me", auth.Me)
			authTenant.PUT("/auth/password", auth.ChangePassword)
			authTenant.GET("/billing/status", billing.GetBillingStatus)
			authTenant.GET("/billing/plans", billing.GetPlans)
			authTenant.POST("/billing/subscribe", billing.Subscribe)
			authTenant.POST("/billing/renew", billing.Renew)
			authTenant.GET("/billing/transactions", billing.GetTransactions)
		}

		// Auth + Tenant + Subscription (Full access)
		fullAccess := api.Group("")
		fullAccess.Use(middleware.AuthMiddleware(), middleware.TenantMiddleware(), middleware.SubscriptionMiddleware())
		{
			fullAccess.GET("/ping", func(c *gin.Context) {
				utils.OK(c, gin.H{"pong": true})
			})

			// Dashboard
			fullAccess.GET("/dashboard", dashboard.GetStats)

			// Notifications
			fullAccess.GET("/notifications", notification.List)
			fullAccess.POST("/notifications/:id/read", notification.MarkAsRead)
			fullAccess.POST("/notifications/read-all", notification.MarkAllAsRead)

			// Audit Logs — denetim kayıtları yalnızca yöneticiye açık.
			fullAccess.GET("/audit-logs", middleware.RequireRole("admin", "superadmin"), audit.List)

			// Reports
			reportsGroup := fullAccess.Group("/reports")
			{
				reportsGroup.GET("/:type", middleware.RequireModulePermission("reports", "read"), report.GetReport)
			}

			// Cari routes
			carisGroup := fullAccess.Group("/caris")
			{
				carisGroup.GET("/summary", middleware.RequireModulePermission("caris", "read"), cari.GetSummary) // Global Summary MUST be before /:id wildcard!
				carisGroup.GET("/next-code", middleware.RequireModulePermission("caris", "read"), cari.GetNextCode)
				carisGroup.GET("", middleware.RequireModulePermission("caris", "read"), cari.List)
				carisGroup.GET("/:id/summary", middleware.RequireModulePermission("caris", "read"), cari.GetCariFinancialSummary)
				carisGroup.GET("/:id", middleware.RequireModulePermission("caris", "read"), cari.GetByID)
				carisGroup.POST("", middleware.RequireModulePermission("caris", "create"), cari.Create)
				carisGroup.PUT("/:id", middleware.RequireModulePermission("caris", "update"), cari.Update)
				carisGroup.DELETE("/:id", middleware.RequireModulePermission("caris", "delete"), cari.Delete)
				carisGroup.GET("/:id/transactions", middleware.RequireModulePermission("caris", "read"), cari.GetTransactions)
				carisGroup.POST("/:id/persons", middleware.RequireModulePermission("caris", "update"), cari.AddPerson)
				carisGroup.PUT("/:id/persons/:person_id", middleware.RequireModulePermission("caris", "update"), cari.UpdatePerson)
				carisGroup.DELETE("/:id/persons/:person_id", middleware.RequireModulePermission("caris", "update"), cari.RemovePerson)
			}

			// Invoice routes
			invoicesGroup := fullAccess.Group("/invoices")
			{
				invoicesGroup.GET("", middleware.RequireModulePermission("invoices", "read"), invoice.List)
				invoicesGroup.GET("/:id", middleware.RequireModulePermission("invoices", "read"), invoice.GetByID)
				invoicesGroup.POST("", middleware.RequireModulePermission("invoices", "create"), invoice.Create)
				invoicesGroup.PUT("/:id", middleware.RequireModulePermission("invoices", "update"), invoice.Update)
				invoicesGroup.PUT("/:id/status", middleware.RequireModulePermission("invoices", "update"), invoice.UpdateStatus)
				invoicesGroup.DELETE("/:id", middleware.RequireModulePermission("invoices", "delete"), invoice.Delete)
				invoicesGroup.POST("/bulk-send", middleware.RequireModulePermission("invoices", "update"), invoice.BulkSend)
				invoicesGroup.POST("/:id/cancel", middleware.RequireModulePermission("invoices", "update"), invoice.Cancel)
				invoicesGroup.POST("/:id/send", middleware.RequireModulePermission("invoices", "update"), invoice.Send)
			}

			// Payment routes
			paymentsGroup := fullAccess.Group("/payments")
			{
				paymentsGroup.GET("", middleware.RequireModulePermission("payments", "read"), payment.List)
				paymentsGroup.GET("/:id", middleware.RequireModulePermission("payments", "read"), payment.GetByID)
				paymentsGroup.POST("", middleware.RequireModulePermission("payments", "create"), payment.Create)
				paymentsGroup.POST("/:id/cancel", middleware.RequireModulePermission("payments", "update"), payment.Cancel)
				paymentsGroup.PUT("/:id", middleware.RequireModulePermission("payments", "update"), payment.Update)
				paymentsGroup.POST("/cash-accounts", middleware.RequireModulePermission("payments", "create"), payment.CreateCashAccount)
				paymentsGroup.GET("/cash-accounts", middleware.RequireModulePermission("payments", "read"), payment.ListCashAccounts)
				paymentsGroup.PUT("/cash-accounts/:id", middleware.RequireModulePermission("payments", "update"), cash.UpdateCashAccount)
				paymentsGroup.DELETE("/cash-accounts/:id", middleware.RequireModulePermission("payments", "delete"), cash.DeleteCashAccount)
				paymentsGroup.POST("/bank-accounts", middleware.RequireModulePermission("payments", "create"), payment.CreateBankAccount)
				paymentsGroup.GET("/bank-accounts", middleware.RequireModulePermission("payments", "read"), payment.ListBankAccounts)
				paymentsGroup.PUT("/bank-accounts/:id", middleware.RequireModulePermission("payments", "update"), cash.UpdateBankAccount)
				paymentsGroup.DELETE("/bank-accounts/:id", middleware.RequireModulePermission("payments", "delete"), cash.DeleteBankAccount)
			}

			// Cash Transaction (Kasa/Banka) routes — "payments" modülünün parçası.
			cashTransactionsGroup := fullAccess.Group("/cash-transactions")
			{
				cashTransactionsGroup.POST("/transfer", middleware.RequireModulePermission("payments", "update"), cash.Transfer)
				cashTransactionsGroup.GET("", middleware.RequireModulePermission("payments", "read"), cash.ListTransactions)
			}

			// Expense routes
			expensesGroup := fullAccess.Group("/expenses")
			{
				expensesGroup.GET("/repeat-analysis", middleware.RequireModulePermission("expenses", "read"), expense.GetRepeatAnalysis)
				expensesGroup.GET("", middleware.RequireModulePermission("expenses", "read"), expense.List)
				expensesGroup.GET("/:id", middleware.RequireModulePermission("expenses", "read"), expense.GetByID)
				expensesGroup.POST("", middleware.RequireModulePermission("expenses", "create"), expense.Create)
				expensesGroup.PUT("/:id", middleware.RequireModulePermission("expenses", "update"), expense.Update)
				expensesGroup.POST("/:id/cancel", middleware.RequireModulePermission("expenses", "update"), expense.Cancel)
			}

			// Expense Category routes
			expenseCategoriesGroup := fullAccess.Group("/expense-categories")
			{
				expenseCategoriesGroup.GET("", middleware.RequireModulePermission("expenses", "read"), expense.ListCategories)
				expenseCategoriesGroup.POST("", middleware.RequireModulePermission("expenses", "create"), expense.CreateCategory)
				expenseCategoriesGroup.PUT("/:id", middleware.RequireModulePermission("expenses", "update"), expense.UpdateCategory)
				expenseCategoriesGroup.DELETE("/:id", middleware.RequireModulePermission("expenses", "delete"), expense.DeleteCategory)
			}

			// Product routes
			productsGroup := fullAccess.Group("/products")
			{
				productsGroup.GET("/next-code", middleware.RequireModulePermission("products", "read"), product.GetNextCode)
				productsGroup.GET("/critical-stock", middleware.RequireModulePermission("products", "read"), product.GetCriticalStock)
				productsGroup.GET("", middleware.RequireModulePermission("products", "read"), product.ListProducts)
				productsGroup.GET("/:id", middleware.RequireModulePermission("products", "read"), product.GetProductByID)
				productsGroup.POST("", middleware.RequireModulePermission("products", "create"), product.CreateProduct)
				productsGroup.PUT("/:id", middleware.RequireModulePermission("products", "update"), product.UpdateProduct)
				productsGroup.DELETE("/:id", middleware.RequireModulePermission("products", "delete"), product.DeleteProduct)
				productsGroup.GET("/:id/movements", middleware.RequireModulePermission("products", "read"), product.ListMovements)
			}

			// Product Category routes
			productCategoriesGroup := fullAccess.Group("/product-categories")
			{
				productCategoriesGroup.GET("", middleware.RequireModulePermission("products", "read"), product.ListCategories)
				productCategoriesGroup.POST("", middleware.RequireModulePermission("products", "create"), product.CreateCategory)
				productCategoriesGroup.PUT("/:id", middleware.RequireModulePermission("products", "update"), product.UpdateCategory)
				productCategoriesGroup.DELETE("/:id", middleware.RequireModulePermission("products", "delete"), product.DeleteCategory)
			}

			// Warehouse routes
			warehousesGroup := fullAccess.Group("/warehouses")
			{
				warehousesGroup.GET("", middleware.RequireModulePermission("products", "read"), product.ListWarehouses)
				warehousesGroup.POST("", middleware.RequireModulePermission("products", "create"), product.CreateWarehouse)
				warehousesGroup.PUT("/:id", middleware.RequireModulePermission("products", "update"), product.UpdateWarehouse)
				warehousesGroup.DELETE("/:id", middleware.RequireModulePermission("products", "delete"), product.DeleteWarehouse)
			}

			// Stock Movement routes
			stockMovementsGroup := fullAccess.Group("/stock-movements")
			{
				stockMovementsGroup.POST("", middleware.RequireModulePermission("products", "update"), product.ManualAdjustment)
			}

			// Quote (Teklif) routes — fatura iş akışının parçası olduğundan
			// "invoices" modül izinleriyle korunur.
			quotesGroup := fullAccess.Group("/quotes")
			{
				quotesGroup.GET("", middleware.RequireModulePermission("invoices", "read"), quote.List)
				quotesGroup.GET("/:id", middleware.RequireModulePermission("invoices", "read"), quote.GetByID)
				quotesGroup.POST("", middleware.RequireModulePermission("invoices", "create"), quote.Create)
				quotesGroup.PUT("/:id", middleware.RequireModulePermission("invoices", "update"), quote.Update)
				quotesGroup.DELETE("/:id", middleware.RequireModulePermission("invoices", "delete"), quote.Delete)
				quotesGroup.POST("/bulk-send", middleware.RequireModulePermission("invoices", "update"), quote.BulkSend)
				quotesGroup.PUT("/:id/status", middleware.RequireModulePermission("invoices", "update"), quote.UpdateStatus)
			quotesGroup.POST("/:id/convert", middleware.RequireModulePermission("invoices", "create"), quote.Convert)
			quotesGroup.POST("/:id/send", middleware.RequireModulePermission("invoices", "update"), quote.Send)
		}

		// Project (Proje) routes — proje yönetimi kendi modül izniyle korunur.
		projectsGroup := fullAccess.Group("/projects")
		{
			projectsGroup.GET("", middleware.RequireModulePermission("projects", "read"), project.List)
			projectsGroup.GET("/:id", middleware.RequireModulePermission("projects", "read"), project.GetByID)
			projectsGroup.POST("", middleware.RequireModulePermission("projects", "create"), project.Create)
			projectsGroup.PUT("/:id", middleware.RequireModulePermission("projects", "update"), project.Update)
			projectsGroup.DELETE("/:id", middleware.RequireModulePermission("projects", "delete"), project.Delete)
			projectsGroup.POST("/:id/invoices/add", middleware.RequireModulePermission("projects", "update"), project.AddInvoice)
			projectsGroup.POST("/:id/invoices/remove", middleware.RequireModulePermission("projects", "update"), project.RemoveInvoice)
			projectsGroup.POST("/:id/quotes/add", middleware.RequireModulePermission("projects", "update"), project.AddQuote)
			projectsGroup.POST("/:id/quotes/remove", middleware.RequireModulePermission("projects", "update"), project.RemoveQuote)
			projectsGroup.POST("/:id/employees/add", middleware.RequireModulePermission("projects", "update"), project.AddEmployee)
			projectsGroup.POST("/:id/employees/remove", middleware.RequireModulePermission("projects", "update"), project.RemoveEmployee)
		}

		// Project Categories (Proje Kategorileri) routes
		projectCategoriesGroup := fullAccess.Group("/project-categories")
		{
			projectCategoriesGroup.GET("", middleware.RequireModulePermission("projects", "read"), projectCategory.List)
			projectCategoriesGroup.GET("/:id", middleware.RequireModulePermission("projects", "read"), projectCategory.GetByID)
			projectCategoriesGroup.POST("", middleware.RequireModulePermission("projects", "create"), projectCategory.Create)
			projectCategoriesGroup.PUT("/:id", middleware.RequireModulePermission("projects", "update"), projectCategory.Update)
			projectCategoriesGroup.DELETE("/:id", middleware.RequireModulePermission("projects", "delete"), projectCategory.Delete)
		}

			// Employee (Personel) routes — personel oluşturma/rol atama yetki
			// yükseltmesine yol açabileceğinden sadece admin/superadmin erişebilir.
			employeesGroup := fullAccess.Group("/employees")
			employeesGroup.Use(middleware.RequireRole("admin", "superadmin"))
			{
				employeesGroup.GET("", employee.List)
				employeesGroup.GET("/:id", employee.GetByID)
				employeesGroup.POST("", employee.Create)
				employeesGroup.PUT("/:id", employee.Update)
				employeesGroup.DELETE("/:id", employee.Delete)
			}

			// Role (Yetki Şablonu) routes — sadece admin/superadmin oluşturup düzenleyebilir
			rolesGroup := fullAccess.Group("/roles")
			rolesGroup.Use(middleware.RequireRole("admin", "superadmin"))
			{
				rolesGroup.GET("", role.List)
				rolesGroup.GET("/:id", role.GetByID)
				rolesGroup.POST("", role.Create)
				rolesGroup.PUT("/:id", role.Update)
				rolesGroup.DELETE("/:id", role.Delete)
			}

			// Settings (Ayarlar) & Company routes.
			// Okuma (profil + ayar anahtarları) fatura/cari/gider formlarında ve
			// yazdırmada gerektiği için tüm girişli kullanıcılara açık.
			// Değişiklik (profil, ayar kaydı, aktif modüller, içe aktarma)
			// yalnızca yöneticiye açık.
			fullAccess.GET("/company", settings.GetCompanyProfile)
			fullAccess.PUT("/company", middleware.RequireRole("admin", "superadmin"), settings.UpdateCompanyProfile)

			settingsGroup := fullAccess.Group("/settings")
			{
				settingsGroup.GET("", settings.ListSettings)
				settingsGroup.GET("/:key", settings.GetSetting)
				settingsGroup.POST("", middleware.RequireRole("admin", "superadmin"), settings.SaveSetting)
				settingsGroup.PUT("/enabled-modules", middleware.RequireRole("admin", "superadmin"), settings.UpdateEnabledModules)

				// Import routes — yalnızca yöneticiye açık.
				settingsGroup.POST("/import/caris", middleware.RequireRole("admin", "superadmin"), importHndlr.ImportCaris)
				settingsGroup.POST("/import/products", middleware.RequireRole("admin", "superadmin"), importHndlr.ImportProducts)
				settingsGroup.GET("/import/sample/caris", middleware.RequireRole("admin", "superadmin"), importHndlr.DownloadSampleCari)
				settingsGroup.GET("/import/sample/products", middleware.RequireRole("admin", "superadmin"), importHndlr.DownloadSampleProduct)
			}

			// Currency (Para Birimi) routes — liste fatura/teklif/cari
			// formlarında gerektiği için tüm girişli kullanıcılara açık;
			// değişiklik (ekle/düzenle/sil) yalnızca yöneticiye açık.
			currenciesGroup := fullAccess.Group("/currencies")
			{
				currenciesGroup.GET("", currency.List)
				currenciesGroup.POST("", middleware.RequireRole("admin", "superadmin"), currency.Create)
				currenciesGroup.PUT("/:id", middleware.RequireRole("admin", "superadmin"), currency.Update)
				currenciesGroup.DELETE("/:id", middleware.RequireRole("admin", "superadmin"), currency.Delete)
			}
		}

		// Superadmin routes (Auth + Role: superadmin, NO TenantMiddleware)
		superadminSvc := services.NewSuperadminService()
		superadmin := handlers.NewSuperadminHandler(superadminSvc)

		announcementSvc := services.NewAnnouncementService()
		announcement := handlers.NewAnnouncementHandler(announcementSvc)

		// Tenant announcements
		authTenant.GET("/announcements", announcement.ListForTenant)
		authTenant.GET("/announcement-categories", announcement.ListCategories)

		superGroup := api.Group("/superadmin")
		superGroup.Use(middleware.AuthMiddleware(), middleware.RequireRole("superadmin"))
		{
			// Dashboard
			superGroup.GET("/dashboard", superadmin.GetDashboardStats)

			// Announcements
			superGroup.GET("/announcements", announcement.ListAll)
			superGroup.POST("/announcements", announcement.Create)
			superGroup.DELETE("/announcements/:id", announcement.Delete)

			// Announcement categories (superadmin yönetir)
			superGroup.GET("/announcement-categories", announcement.ListCategories)
			superGroup.POST("/announcement-categories", announcement.CreateCategory)
			superGroup.DELETE("/announcement-categories/:id", announcement.DeleteCategory)

			// Company Management
			superGroup.GET("/companies", superadmin.GetCompanies)
			superGroup.POST("/companies", superadmin.CreateCompany)
			superGroup.PUT("/companies/:id", superadmin.UpdateCompany)
			superGroup.DELETE("/companies/:id", superadmin.DeleteCompany)
			superGroup.PUT("/companies/:id/status", superadmin.ToggleCompanyStatus)

			// Plans Management
			superGroup.GET("/plans", superadmin.GetPlans)
			superGroup.POST("/plans", superadmin.CreatePlan)
			superGroup.PUT("/plans/:id", superadmin.UpdatePlan)
			superGroup.DELETE("/plans/:id", superadmin.DeletePlan)

			// Email (SMTP) settings - platform-wide
			superGroup.GET("/email-settings", superadmin.GetEmailSettings)
			superGroup.PUT("/email-settings", superadmin.UpdateEmailSettings)
			superGroup.POST("/email-settings/test", superadmin.TestEmailSettings)
		}
	}

	return r
}

func CORS() gin.HandlerFunc {
	// İzinli origin listesi: FRONTEND_ORIGIN (virgülle birden fazla verilebilir).
	// Boşsa geliştirme varsayılanları kullanılır.
	raw := os.Getenv("FRONTEND_ORIGIN")
	if raw == "" {
		raw = "http://localhost:9000,http://localhost:5173"
	}
	allowed := map[string]bool{}
	for _, o := range strings.Split(raw, ",") {
		if o = strings.TrimSpace(o); o != "" {
			allowed[o] = true
		}
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		// Sadece whitelist'teki origin'e izin ver (credentialed istek için zorunlu).
		if origin != "" && allowed[origin] {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Vary", "Origin")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Cookie")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
