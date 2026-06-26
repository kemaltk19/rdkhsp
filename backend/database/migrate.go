package database

import (
	"errors"
	"log"
	"os"

	"gorm.io/gorm"

	"radikal-hesap/models"
	"radikal-hesap/utils"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Migrate runs schema migrations and enforces PostgreSQL Row Level Security policies.
func Migrate() {
	if SystemDB == nil {
		log.Fatal("[migrate] SystemDB nil, migration baslatilamadi")
	}

	log.Println("[migrate] Veritabani sema migrasyonu baslatiliyor...")
	err := SystemDB.AutoMigrate(
		&models.Plan{},
		&models.Company{},
		&models.User{},
		&models.RefreshToken{},
		&models.VerificationToken{},
		&models.EmailSetting{},
		&models.NumberSequence{},
		&models.Cari{},
		&models.CariPerson{},
		&models.CariBalance{},
		&models.CariTransaction{},
		&models.Invoice{},
		&models.InvoiceItem{},
		&models.CashAccount{},
		&models.BankAccount{},
		&models.CashTransaction{},
		&models.Payment{},
		&models.ExpenseCategory{},
		&models.Expense{},
		&models.ProductCategory{},
		&models.Product{},
		&models.Warehouse{},
		&models.StockMovement{},
		&models.Quote{},
		&models.QuoteItem{},
		&models.Project{},
		&models.ProjectCategory{},
		&models.Employee{},
		&models.Role{},
		&models.RolePermission{},
		&models.Setting{},
		&models.Currency{},
		&models.AuditLog{},
		&models.Announcement{},
		&models.AnnouncementCategory{},
		&models.BillingTransaction{},
		&models.Notification{},
	)
	if err != nil {
		log.Fatalf("[migrate] AutoMigrate hatasi: %v", err)
	}
	log.Println("[migrate] AutoMigrate basariyla tamamlandi")

	log.Println("[migrate] Eski Cari Gruplari ayarlari temizleniyor...")
	SystemDB.Exec("DELETE FROM settings WHERE key = 'cari_groups'")

	log.Println("[migrate] RLS politikalari yapilandiriliyor...")

	// Manuel İdempotent Migrasyonlar
	log.Println("[migrate] Manuel idempotent SQL migrasyonlari calistiriliyor...")
	manualSQLs := []string{
		`ALTER TABLE invoices ALTER COLUMN exchange_rate TYPE numeric(65,10);`,
		`ALTER TABLE products ALTER COLUMN min_stock SET DEFAULT 5;`,
		`UPDATE products SET average_cost = purchase_price WHERE average_cost = 0 AND purchase_price > 0 AND type = 'product';`,
		`UPDATE announcements SET category = 'bilgi' WHERE category IS NULL OR category = '';`,
		// invoice_prefix tek anahtari satış/alış için paylasiliyor, artik ayri key'lere taşındi.
		// Eski kaydi sil (varsa) — yeni seed fonksiyonu ayrı key'ler ekler.
		`DELETE FROM settings WHERE key = 'invoice_prefix';`,
	}
	for _, sql := range manualSQLs {
		if err := SystemDB.Exec(sql).Error; err != nil {
			log.Printf("[migrate] Manuel SQL uyarisi (%s): %v", sql, err)
		}
	}

	rlsSQLs := []string{
		`ALTER TABLE companies ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE companies FORCE ROW LEVEL SECURITY;`,
		`ALTER TABLE users ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE users FORCE ROW LEVEL SECURITY;`,
		`ALTER TABLE number_sequences ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE number_sequences FORCE ROW LEVEL SECURITY;`,
		`ALTER TABLE caris ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE caris FORCE ROW LEVEL SECURITY;`,
		`ALTER TABLE cari_persons ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE cari_persons FORCE ROW LEVEL SECURITY;`,
		`ALTER TABLE cari_balances ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE cari_balances FORCE ROW LEVEL SECURITY;`,
		`ALTER TABLE cari_transactions ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE cari_transactions FORCE ROW LEVEL SECURITY;`,
		`ALTER TABLE invoices ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE invoices FORCE ROW LEVEL SECURITY;`,
		`ALTER TABLE invoice_items ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE invoice_items FORCE ROW LEVEL SECURITY;`,
		`ALTER TABLE cash_accounts ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE cash_accounts FORCE ROW LEVEL SECURITY;`,
		`ALTER TABLE bank_accounts ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE bank_accounts FORCE ROW LEVEL SECURITY;`,
		`ALTER TABLE cash_transactions ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE cash_transactions FORCE ROW LEVEL SECURITY;`,
		`ALTER TABLE payments ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE payments FORCE ROW LEVEL SECURITY;`,
		`ALTER TABLE expense_categories ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE expense_categories FORCE ROW LEVEL SECURITY;`,
		`ALTER TABLE expenses ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE expenses FORCE ROW LEVEL SECURITY;`,
		`ALTER TABLE product_categories ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE product_categories FORCE ROW LEVEL SECURITY;`,
		`ALTER TABLE products ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE products FORCE ROW LEVEL SECURITY;`,
		`ALTER TABLE warehouses ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE warehouses FORCE ROW LEVEL SECURITY;`,
		`ALTER TABLE stock_movements ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE stock_movements FORCE ROW LEVEL SECURITY;`,
		`ALTER TABLE quotes ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE quotes FORCE ROW LEVEL SECURITY;`,
		`ALTER TABLE quote_items ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE quote_items FORCE ROW LEVEL SECURITY;`,
		`ALTER TABLE employees ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE employees FORCE ROW LEVEL SECURITY;`,
		`ALTER TABLE settings ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE settings FORCE ROW LEVEL SECURITY;`,
		`ALTER TABLE currencies ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE currencies FORCE ROW LEVEL SECURITY;`,
		`ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE audit_logs FORCE ROW LEVEL SECURITY;`,
		`ALTER TABLE billing_transactions ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE billing_transactions FORCE ROW LEVEL SECURITY;`,
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'billing_transactions' AND policyname = 'billing_tx_tenant_policy') THEN
				CREATE POLICY billing_tx_tenant_policy ON billing_transactions
				USING (company_id = NULLIF(current_setting('app.company_id', true), '')::uuid);
			END IF;
		END
		$$;`,

		// Company policy: must match current_setting('app.company_id')
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'companies' AND policyname = 'company_tenant_policy') THEN
				CREATE POLICY company_tenant_policy ON companies
				USING (id = NULLIF(current_setting('app.company_id', true), '')::uuid);
			END IF;
		END
		$$;`,

		// User policy: must match current_setting('app.company_id')
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'users' AND policyname = 'user_tenant_policy') THEN
				CREATE POLICY user_tenant_policy ON users
				USING (company_id = NULLIF(current_setting('app.company_id', true), '')::uuid);
			END IF;
		END
		$$;`,

		// Number sequences policy: must match current_setting('app.company_id')
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'number_sequences' AND policyname = 'numseq_tenant_policy') THEN
				CREATE POLICY numseq_tenant_policy ON number_sequences
				USING (company_id = NULLIF(current_setting('app.company_id', true), '')::uuid);
			END IF;
		END
		$$;`,

		// Cari policy: must match current_setting('app.company_id')
		`DO $$
		BEGIN
			DROP POLICY IF EXISTS "company_isolation_caris" ON caris;
			CREATE POLICY "company_isolation_caris" ON caris FOR ALL USING (company_id = NULLIF(current_setting('app.company_id', true), '')::uuid);
			
			DROP POLICY IF EXISTS "company_isolation_cari_persons" ON cari_persons;
			CREATE POLICY "company_isolation_cari_persons" ON cari_persons FOR ALL USING (cari_id IN (SELECT id FROM caris WHERE company_id = NULLIF(current_setting('app.company_id', true), '')::uuid));
			
			DROP POLICY IF EXISTS "company_isolation_cari_balances" ON cari_balances;
			CREATE POLICY "company_isolation_cari_balances" ON cari_balances FOR ALL USING (cari_id IN (SELECT id FROM caris WHERE company_id = NULLIF(current_setting('app.company_id', true), '')::uuid));
		END
		$$;`,

		// CariTransaction policy: must match current_setting('app.company_id')
		`DO $$
		BEGIN
			DROP POLICY IF EXISTS "company_isolation_cari_transactions" ON cari_transactions;
			CREATE POLICY "company_isolation_cari_transactions" ON cari_transactions FOR ALL USING (company_id = NULLIF(current_setting('app.company_id', true), '')::uuid);
		END
		$$;`,

		// Invoice policy: must match current_setting('app.company_id')
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'invoices' AND policyname = 'invoice_tenant_policy') THEN
				CREATE POLICY invoice_tenant_policy ON invoices
				USING (company_id = NULLIF(current_setting('app.company_id', true), '')::uuid);
			END IF;
		END
		$$;`,

		// InvoiceItem policy: must match current_setting('app.company_id')
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'invoice_items' AND policyname = 'invoice_item_tenant_policy') THEN
				CREATE POLICY invoice_item_tenant_policy ON invoice_items
				USING (company_id = NULLIF(current_setting('app.company_id', true), '')::uuid);
			END IF;
		END
		$$;`,

		// CashAccount policy: must match current_setting('app.company_id')
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'cash_accounts' AND policyname = 'cash_acc_tenant_policy') THEN
				CREATE POLICY cash_acc_tenant_policy ON cash_accounts
				USING (company_id = NULLIF(current_setting('app.company_id', true), '')::uuid);
			END IF;
		END
		$$;`,

		// BankAccount policy: must match current_setting('app.company_id')
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'bank_accounts' AND policyname = 'bank_acc_tenant_policy') THEN
				CREATE POLICY bank_acc_tenant_policy ON bank_accounts
				USING (company_id = NULLIF(current_setting('app.company_id', true), '')::uuid);
			END IF;
		END
		$$;`,

		// CashTransaction policy: must match current_setting('app.company_id')
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'cash_transactions' AND policyname = 'cash_tx_tenant_policy') THEN
				CREATE POLICY cash_tx_tenant_policy ON cash_transactions
				USING (company_id = NULLIF(current_setting('app.company_id', true), '')::uuid);
			END IF;
		END
		$$;`,

		// Payment policy: must match current_setting('app.company_id')
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'payments' AND policyname = 'payment_tenant_policy') THEN
				CREATE POLICY payment_tenant_policy ON payments
				USING (company_id = NULLIF(current_setting('app.company_id', true), '')::uuid);
			END IF;
		END
		$$;`,

		// ExpenseCategory policy: must match current_setting('app.company_id')
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'expense_categories' AND policyname = 'exp_cat_tenant_policy') THEN
				CREATE POLICY exp_cat_tenant_policy ON expense_categories
				USING (company_id = NULLIF(current_setting('app.company_id', true), '')::uuid);
			END IF;
		END
		$$;`,

		// Expense policy: must match current_setting('app.company_id')
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'expenses' AND policyname = 'expense_tenant_policy') THEN
				CREATE POLICY expense_tenant_policy ON expenses
				USING (company_id = NULLIF(current_setting('app.company_id', true), '')::uuid);
			END IF;
		END
		$$;`,

		// ProductCategory policy: must match current_setting('app.company_id')
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'product_categories' AND policyname = 'prod_cat_tenant_policy') THEN
				CREATE POLICY prod_cat_tenant_policy ON product_categories
				USING (company_id = NULLIF(current_setting('app.company_id', true), '')::uuid);
			END IF;
		END
		$$;`,

		// Product policy: must match current_setting('app.company_id')
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'products' AND policyname = 'product_tenant_policy') THEN
				CREATE POLICY product_tenant_policy ON products
				USING (company_id = NULLIF(current_setting('app.company_id', true), '')::uuid);
			END IF;
		END
		$$;`,

		// Warehouse policy: must match current_setting('app.company_id')
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'warehouses' AND policyname = 'warehouse_tenant_policy') THEN
				CREATE POLICY warehouse_tenant_policy ON warehouses
				USING (company_id = NULLIF(current_setting('app.company_id', true), '')::uuid);
			END IF;
		END
		$$;`,

		// StockMovement policy: must match current_setting('app.company_id')
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'stock_movements' AND policyname = 'stock_mov_tenant_policy') THEN
				CREATE POLICY stock_mov_tenant_policy ON stock_movements
				USING (company_id = NULLIF(current_setting('app.company_id', true), '')::uuid);
			END IF;
		END
		$$;`,

		// Quote policy: must match current_setting('app.company_id')
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'quotes' AND policyname = 'quote_tenant_policy') THEN
				CREATE POLICY quote_tenant_policy ON quotes
				USING (company_id = NULLIF(current_setting('app.company_id', true), '')::uuid);
			END IF;
		END
		$$;`,

		// QuoteItem policy: must match current_setting('app.company_id')
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'quote_items' AND policyname = 'quote_item_tenant_policy') THEN
				CREATE POLICY quote_item_tenant_policy ON quote_items
				USING (company_id = NULLIF(current_setting('app.company_id', true), '')::uuid);
			END IF;
		END
		$$;`,

		// Project RLS
		`ALTER TABLE projects ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE projects FORCE ROW LEVEL SECURITY;`,
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'projects' AND policyname = 'project_tenant_policy') THEN
				CREATE POLICY project_tenant_policy ON projects
				USING (company_id = NULLIF(current_setting('app.company_id', true), '')::uuid);
			END IF;
		END
		$$;`,

		// ProjectCategory RLS
		`ALTER TABLE project_categories ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE project_categories FORCE ROW LEVEL SECURITY;`,
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'project_categories' AND policyname = 'project_category_tenant_policy') THEN
				CREATE POLICY project_category_tenant_policy ON project_categories
				USING (company_id = NULLIF(current_setting('app.company_id', true), '')::uuid);
			END IF;
		END
		$$;`,

		// Currency policy: must match current_setting('app.company_id')
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'currencies' AND policyname = 'currency_tenant_policy') THEN
				CREATE POLICY currency_tenant_policy ON currencies
				USING (company_id = NULLIF(current_setting('app.company_id', true), '')::uuid);
			END IF;
		END
		$$;`,

		// Unique company_id + code index on caris
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_caris_company_code ON caris (company_id, code);`,

		// Unique company_id + number index on invoices
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_invoices_company_number ON invoices (company_id, number);`,

		// Unique company_id + code index on quotes
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_quotes_company_number ON quotes (company_id, number);`,

		// Unique company_id + code index on projects
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_projects_company_code ON projects (company_id, code);`,

		// Unique company_id + code index on project_categories
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_project_categories_company_code ON project_categories (company_id, code);`,

		// Composite index on cari_transactions
		`CREATE INDEX IF NOT EXISTS idx_cari_transactions_comp_cari_date ON cari_transactions (company_id, cari_id, date);`,

		// Composite index on invoices
		`CREATE INDEX IF NOT EXISTS idx_invoices_comp_cari_date ON invoices (company_id, cari_id, date);`,

		// Composite index on quotes
		`CREATE INDEX IF NOT EXISTS idx_quotes_comp_cari_date ON quotes (company_id, cari_id, date);`,

		// Composite index on cash_transactions
		`CREATE INDEX IF NOT EXISTS idx_cash_transactions_comp_kind_id_date ON cash_transactions (company_id, account_kind, account_id, date);`,

		// Composite index on payments
		`CREATE INDEX IF NOT EXISTS idx_payments_comp_cari_date ON payments (company_id, cari_id, date);`,

		// Composite index on expenses
		`CREATE INDEX IF NOT EXISTS idx_expenses_comp_cat_cari_date ON expenses (company_id, category_id, cari_id, date);`,

		// Unique company_id + code index on products
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_products_company_code ON products (company_id, code);`,

		// Composite index on stock_movements
		`CREATE INDEX IF NOT EXISTS idx_stock_movements_comp_prod_wh_date ON stock_movements (company_id, product_id, warehouse_id, date);`,

		// Employees RLS
		`ALTER TABLE employees ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE employees FORCE ROW LEVEL SECURITY;`,
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'employees' AND policyname = 'employee_tenant_policy') THEN
				CREATE POLICY employee_tenant_policy ON employees
				USING (company_id = NULLIF(current_setting('app.company_id', true), '')::uuid);
			END IF;
		END
		$$;`,

		// Roles RLS
		`ALTER TABLE roles ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE roles FORCE ROW LEVEL SECURITY;`,
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'roles' AND policyname = 'role_tenant_policy') THEN
				CREATE POLICY role_tenant_policy ON roles
				USING (company_id = NULLIF(current_setting('app.company_id', true), '')::uuid);
			END IF;
		END
		$$;`,

		// RolePermissions RLS — role_permissions'ta company_id yok, role_id üzerinden join ile filtrelenir (cari_persons deseni)
		`ALTER TABLE role_permissions ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE role_permissions FORCE ROW LEVEL SECURITY;`,
		`DO $$
		BEGIN
			DROP POLICY IF EXISTS "company_isolation_role_permissions" ON role_permissions;
			CREATE POLICY "company_isolation_role_permissions" ON role_permissions FOR ALL
			USING (role_id IN (SELECT id FROM roles WHERE company_id = NULLIF(current_setting('app.company_id', true), '')::uuid));
		END
		$$;`,

		// Bir role içinde aynı modül için tekrar satır olmasın
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_role_permissions_role_module ON role_permissions (role_id, module);`,

		// User.RoleID -> roles.id FK (rol silinirse kullanıcı izinsiz kalır ama bozulmaz)
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_users_role_id') THEN
				ALTER TABLE users ADD CONSTRAINT fk_users_role_id FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE SET NULL;
			END IF;
		END
		$$;`,

		// Settings RLS
		`ALTER TABLE settings ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE settings FORCE ROW LEVEL SECURITY;`,
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'settings' AND policyname = 'settings_tenant_policy') THEN
				CREATE POLICY settings_tenant_policy ON settings
				USING (company_id = NULLIF(current_setting('app.company_id', true), '')::uuid);
			END IF;
		END
		$$;`,

		// AuditLogs RLS
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'audit_logs' AND policyname = 'audit_log_tenant_policy') THEN
				CREATE POLICY audit_log_tenant_policy ON audit_logs
				USING (company_id = NULLIF(current_setting('app.company_id', true), '')::uuid);
			END IF;
		END
		$$;`,

		// Announcements RLS (tenants can read all announcements, superadmin bypasses)
		`ALTER TABLE announcements ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE announcements FORCE ROW LEVEL SECURITY;`,
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'announcements' AND policyname = 'announcements_read_policy') THEN
				CREATE POLICY announcements_read_policy ON announcements
				FOR SELECT
				USING (true);
			END IF;
		END
		$$;`,

		// Notifications RLS
		`ALTER TABLE notifications ENABLE ROW LEVEL SECURITY;`,
		`ALTER TABLE notifications FORCE ROW LEVEL SECURITY;`,
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'notifications' AND policyname = 'notification_tenant_policy') THEN
				CREATE POLICY notification_tenant_policy ON notifications
				USING (company_id = NULLIF(current_setting('app.company_id', true), '')::uuid);
			END IF;
		END
		$$;`,
	}

	for _, sql := range rlsSQLs {
		if err := SystemDB.Exec(sql).Error; err != nil {
			log.Fatalf("[migrate] RLS SQL hatasi (%s): %v", sql, err)
		}
	}
	log.Println("[migrate] RLS politikalari basariyla uygulandi")

	// Grant table/sequence access to the runtime app role. AutoMigrate may create
	// new tables (e.g. currencies) that radikal_app cannot access by default,
	// which causes 'permission denied' (SQLSTATE 42501) at runtime. Re-grant here.
	log.Println("[migrate] App rol yetkileri (GRANT) uygulaniyor...")
	grantSQLs := []string{
		`GRANT USAGE ON SCHEMA public TO radikal_app, radikal_system`,
		`GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO radikal_app, radikal_system`,
		`GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO radikal_app, radikal_system`,
		`ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON TABLES TO radikal_app, radikal_system`,
		`ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON SEQUENCES TO radikal_app, radikal_system`,
	}
	for _, sql := range grantSQLs {
		if err := SystemDB.Exec(sql).Error; err != nil {
			log.Printf("[migrate] GRANT uyarisi (%s): %v", sql, err)
		}
	}
	log.Println("[migrate] App rol yetkileri uygulandi")

	log.Println("[migrate] Plan tohumlamasi yapiliyor...")
	seedPlans()

	log.Println("[migrate] Superadmin tohumlamasi yapiliyor...")
	seedSuperAdmin()

	log.Println("[migrate] Varsayilan urun kategorileri tohumlamasi yapiliyor...")
	seedDefaultCategories()

	log.Println("[migrate] Varsayilan ayarlar (Cari gruplari, Teklif ön eki vs) tohumlamasi yapiliyor...")
	seedDefaultSettings()

	log.Println("[migrate] Varsayilan gider kategorileri tohumlamasi yapiliyor...")
	seedDefaultExpenseCategories()

	log.Println("[migrate] Varsayilan kasa ve banka hesaplari tohumlamasi yapiliyor...")
	seedDefaultAccounts()
	seedDefaultCurrencies()
	log.Println("[migrate] Varsayilan yetki rolleri tohumlamasi yapiliyor...")
	seedDefaultRoles()

	log.Println("[migrate] Varsayilan duyuru kategorileri tohumlamasi yapiliyor...")
	seedAnnouncementCategories()
}

// seedAnnouncementCategories, varsayılan duyuru kategorilerini ekler. İdempotent:
// her slug yalnızca yoksa eklenir; superadmin sonradan ekleyip silebilir.
func seedAnnouncementCategories() {
	defaults := []models.AnnouncementCategory{
		{Slug: "bilgi", Name: "Bilgi"},
		{Slug: "egitim", Name: "Eğitim"},
		{Slug: "hata", Name: "Hata"},
		{Slug: "ozellik", Name: "Özellik"},
	}
	for _, d := range defaults {
		var count int64
		SystemDB.Model(&models.AnnouncementCategory{}).Where("slug = ?", d.Slug).Count(&count)
		if count > 0 {
			continue
		}
		d.ID = uuid.New()
		if err := SystemDB.Create(&d).Error; err != nil {
			log.Printf("[migrate] Duyuru kategorisi tohumlama uyarisi (%s): %v", d.Slug, err)
		}
	}
}

func seedPlans() {
	var count int64
	SystemDB.Model(&models.Plan{}).Count(&count)
	if count > 0 {
		return
	}

	// Create Plans
	freeID, _ := uuid.Parse("00000000-0000-0000-0000-000000000001")
	stdID, _ := uuid.Parse("00000000-0000-0000-0000-000000000002")
	proID, _ := uuid.Parse("00000000-0000-0000-0000-000000000003")

	plans := []models.Plan{
		{
			ID:           freeID,
			Name:         "Deneme / Ücretsiz",
			Code:         "free",
			PriceMonthly: decimal.NewFromFloat(0.00),
			PriceYearly:  decimal.NewFromFloat(0.00),
			Currency:     "TRY",
			Features:     `["dashboard", "caris", "invoices", "payments", "expenses", "products", "reports"]`,
			IsActive:     true,
		},
		{
			ID:           stdID,
			Name:         "Standart Paket",
			Code:         "standard",
			PriceMonthly: decimal.NewFromFloat(199.00),
			PriceYearly:  decimal.NewFromFloat(1990.00),
			Currency:     "TRY",
			Features:     `["dashboard", "caris", "invoices", "payments", "expenses", "products", "reports", "employees", "settings"]`,
			IsActive:     true,
		},
		{
			ID:           proID,
			Name:         "Profesyonel Paket",
			Code:         "pro",
			PriceMonthly: decimal.NewFromFloat(499.00),
			PriceYearly:  decimal.NewFromFloat(4990.00),
			Currency:     "TRY",
			Features:     `["dashboard", "caris", "invoices", "payments", "expenses", "products", "reports", "employees", "settings"]`,
			IsActive:     true,
		},
	}

	for _, p := range plans {
		if err := SystemDB.Create(&p).Error; err != nil {
			log.Printf("[migrate] Plan tohumlama hatasi (%s): %v", p.Name, err)
		}
	}
	log.Println("[migrate] Planlar basariyla tohumlandi")
}

func seedSuperAdmin() {
	email := os.Getenv("SUPERADMIN_EMAIL")
	password := os.Getenv("SUPERADMIN_PASSWORD")
	env := os.Getenv("APP_ENV")
	nodeEnv := os.Getenv("NODE_ENV")
	isProd := env == "production" || nodeEnv == "production" || env == "staging" || nodeEnv == "staging"

	if email == "" || password == "" {
		if isProd {
			log.Fatal("[migrate] Production/Staging ortaminda SUPERADMIN_EMAIL ve SUPERADMIN_PASSWORD tanımlanmak zorundadır!")
		}
		if email == "" {
			email = "admin@radikalhesap.local"
		}
		if password == "" {
			password = "superadmin_dev_password_123!"
		}
		log.Printf("[migrate] UYARI: SUPERADMIN bilgileri bos. Varsayilan gelistirme bilgileri kullaniliyor: %s", email)
	}

	var count int64
	SystemDB.Model(&models.User{}).Where("email = ?", email).Count(&count)
	if count > 0 {
		log.Println("[migrate] Superadmin zaten mevcut, atlaniyor")
		return
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		log.Fatalf("[migrate] Superadmin sifre olusturma hatasi: %v", err)
	}

	superAdmin := models.User{
		ID:           uuid.New(),
		CompanyID:    nil,
		Name:         "Super Admin",
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         "superadmin",
		Locale:       "tr",
		IsActive:     true,
	}

	if err := SystemDB.Create(&superAdmin).Error; err != nil {
		log.Fatalf("[migrate] Superadmin tohumlama hatasi: %v", err)
	}
	log.Println("[migrate] Superadmin basariyla tohumlandi")
}

func seedDefaultCategories() {
	var companies []models.Company
	if err := SystemDB.Find(&companies).Error; err != nil {
		return
	}
	for _, company := range companies {
		var count int64
		SystemDB.Model(&models.ProductCategory{}).Where("company_id = ?", company.ID).Count(&count)
		if count == 0 {
			categories := []models.ProductCategory{
				{CompanyID: company.ID, Name: "Ticari Mallar"},
				{CompanyID: company.ID, Name: "Hammaddeler"},
				{CompanyID: company.ID, Name: "Yarı Mamuller"},
				{CompanyID: company.ID, Name: "Mamuller"},
				{CompanyID: company.ID, Name: "Sarf Malzemeleri"},
				{CompanyID: company.ID, Name: "Hizmet Ürünleri"},
				{CompanyID: company.ID, Name: "Demirbaşlar"},
			}
			for i := range categories {
				categories[i].ID, _ = uuid.NewV7()
			}
			if err := SystemDB.Create(&categories).Error; err != nil {
				log.Printf("[migrate] Default category tohumlama hatasi (company: %s): %v", company.ID, err)
			} else {
				log.Printf("[migrate] Default category tohumlama basarili (company: %s)", company.ID)
			}
		}
	}
}

func seedDefaultSettings() {
	var companies []models.Company
	if err := SystemDB.Find(&companies).Error; err != nil {
		return
	}
	for _, company := range companies {
		// Cari Groups
		var countCari int64
		SystemDB.Model(&models.Setting{}).Where("company_id = ? AND key = 'cari_groups'", company.ID).Count(&countCari)
		if countCari == 0 {
			defaultGroups := `["Bireysel", "Kurumsal", "Kurum", "Fabrika", "Esnaf", "Şirket", "Diğer"]`
			settingID, _ := uuid.NewV7()
			setting := models.Setting{
				ID:        settingID,
				CompanyID: company.ID,
				Key:       "cari_groups",
				Value:     defaultGroups,
				Category:  "cari",
			}
			if err := SystemDB.Create(&setting).Error; err != nil {
				log.Printf("[migrate] Default cari_groups tohumlama hatasi (company: %s): %v", company.ID, err)
			} else {
				log.Printf("[migrate] Default cari_groups tohumlama basarili (company: %s)", company.ID)
			}
		}

		// Quote Prefix
		var countQuote int64
		SystemDB.Model(&models.Setting{}).Where("company_id = ? AND key = 'quote_prefix'", company.ID).Count(&countQuote)
		if countQuote == 0 {
			settingID, _ := uuid.NewV7()
			setting := models.Setting{
				ID:        settingID,
				CompanyID: company.ID,
				Key:       "quote_prefix",
				Value:     "PRO",
				Category:  "quote",
			}
			if err := SystemDB.Create(&setting).Error; err != nil {
				log.Printf("[migrate] Default quote_prefix tohumlama hatasi (company: %s): %v", company.ID, err)
			} else {
				log.Printf("[migrate] Default quote_prefix tohumlama basarili (company: %s)", company.ID)
			}
		}

		// Invoice Sales Prefix (satış faturası)
		var countInvSales int64
		SystemDB.Model(&models.Setting{}).Where("company_id = ? AND key = 'invoice_sales_prefix'", company.ID).Count(&countInvSales)
		if countInvSales == 0 {
			sid, _ := uuid.NewV7()
			SystemDB.Create(&models.Setting{ID: sid, CompanyID: company.ID, Key: "invoice_sales_prefix", Value: "INV-S", Category: "invoice"})
		}

		// Invoice Purchase Prefix (alış faturası)
		var countInvPurchase int64
		SystemDB.Model(&models.Setting{}).Where("company_id = ? AND key = 'invoice_purchase_prefix'", company.ID).Count(&countInvPurchase)
		if countInvPurchase == 0 {
			pid, _ := uuid.NewV7()
			SystemDB.Create(&models.Setting{ID: pid, CompanyID: company.ID, Key: "invoice_purchase_prefix", Value: "BILL", Category: "invoice"})
		}
	}
}

func seedDefaultExpenseCategories() {
	var companies []models.Company
	if err := SystemDB.Find(&companies).Error; err != nil {
		return
	}
	for _, company := range companies {
		var count int64
		SystemDB.Model(&models.ExpenseCategory{}).Where("company_id = ?", company.ID).Count(&count)
		if count == 0 {
			categories := []models.ExpenseCategory{
				{CompanyID: company.ID, Name: "Elektrik"},
				{CompanyID: company.ID, Name: "Su"},
				{CompanyID: company.ID, Name: "Doğalgaz"},
				{CompanyID: company.ID, Name: "İnternet"},
				{CompanyID: company.ID, Name: "Telefon"},
				{CompanyID: company.ID, Name: "Temizlik"},
				{CompanyID: company.ID, Name: "Kargo"},
				{CompanyID: company.ID, Name: "Yemek"},
			}
			for i := range categories {
				categories[i].ID, _ = uuid.NewV7()
			}
			if err := SystemDB.Create(&categories).Error; err != nil {
				log.Printf("[migrate] Default expense category tohumlama hatasi (company: %s): %v", company.ID, err)
			} else {
				log.Printf("[migrate] Default expense category tohumlama basarili (company: %s)", company.ID)
			}
		}
	}
}

func seedDefaultAccounts() {
	var companies []models.Company
	if err := SystemDB.Find(&companies).Error; err != nil {
		return
	}
	
	cashCurrencies := []struct{ Code, Name, Curr string }{
		{"KAS-TL", "Merkez Kasa (TRY)", "TRY"},
		{"KAS-USD", "Merkez Kasa (USD)", "USD"},
		{"KAS-EUR", "Merkez Kasa (EUR)", "EUR"},
		{"KAS-GBP", "Merkez Kasa (GBP)", "GBP"},
		{"KAS-RUB", "Merkez Kasa (RUB)", "RUB"},
	}
	
	bankCurrencies := []struct{ Code, Name, Curr string }{
		{"BNK-TL", "Banka Hesabı (TRY)", "TRY"},
		{"BNK-USD", "Banka Hesabı (USD)", "USD"},
		{"BNK-EUR", "Banka Hesabı (EUR)", "EUR"},
		{"BNK-GBP", "Banka Hesabı (GBP)", "GBP"},
		{"BNK-RUB", "Banka Hesabı (RUB)", "RUB"},
	}

	for _, company := range companies {
		// Cash Accounts
		for _, c := range cashCurrencies {
			var count int64
			SystemDB.Model(&models.CashAccount{}).Where("company_id = ? AND currency = ?", company.ID, c.Curr).Count(&count)
			if count == 0 {
				id, _ := uuid.NewV7()
				acc := models.CashAccount{
					ID:        id,
					CompanyID: company.ID,
					Code:      c.Code,
					Name:      c.Name,
					Currency:  c.Curr,
					Balance:   decimal.Zero,
					IsDefault: c.Curr == "TRY",
				}
				SystemDB.Create(&acc)
			}
		}

		// Bank Accounts
		for _, b := range bankCurrencies {
			var count int64
			SystemDB.Model(&models.BankAccount{}).Where("company_id = ? AND currency = ?", company.ID, b.Curr).Count(&count)
			if count == 0 {
				id, _ := uuid.NewV7()
				acc := models.BankAccount{
					ID:        id,
					CompanyID: company.ID,
					Code:      b.Code,
					Name:      b.Name,
					Currency:  b.Curr,
					Balance:   decimal.Zero,
				}
				SystemDB.Create(&acc)
			}
		}
	}
}

func seedDefaultCurrencies() {
	var companies []models.Company
	if err := SystemDB.Find(&companies).Error; err != nil {
		return
	}

	currencies := []models.Currency{
		{Name: "Türk Lirası", Symbol: "₺", Code: "TRY", ExchangeRate: decimal.NewFromInt(1), FormatPosition: "RightSpace", FormatThousandSep: ".", FormatDecimalSep: ",", FormatDecimals: 2, IsDefault: true},
		{Name: "Amerikan Doları", Symbol: "$", Code: "USD", ExchangeRate: decimal.NewFromInt(1), FormatPosition: "Left", FormatThousandSep: ",", FormatDecimalSep: ".", FormatDecimals: 2, IsDefault: false},
		{Name: "Euro", Symbol: "€", Code: "EUR", ExchangeRate: decimal.NewFromInt(1), FormatPosition: "RightSpace", FormatThousandSep: ".", FormatDecimalSep: ",", FormatDecimals: 2, IsDefault: false},
		{Name: "İngiliz Sterlini", Symbol: "£", Code: "GBP", ExchangeRate: decimal.NewFromInt(1), FormatPosition: "Left", FormatThousandSep: ",", FormatDecimalSep: ".", FormatDecimals: 2, IsDefault: false},
		{Name: "Rus Rublesi", Symbol: "₽", Code: "RUB", ExchangeRate: decimal.NewFromInt(1), FormatPosition: "RightSpace", FormatThousandSep: ".", FormatDecimalSep: ",", FormatDecimals: 2, IsDefault: false},
	}

	for _, company := range companies {
		for _, cur := range currencies {
			var count int64
			SystemDB.Model(&models.Currency{}).Where("company_id = ? AND code = ?", company.ID, cur.Code).Count(&count)
			if count == 0 {
				newCur := cur
				newCur.ID, _ = uuid.NewV7()
				newCur.CompanyID = company.ID
				SystemDB.Create(&newCur)
			}
		}
	}
}

func seedDefaultRoles() {
	var companies []models.Company
	if err := SystemDB.Find(&companies).Error; err != nil {
		return
	}

	modules := []string{"caris", "invoices", "payments", "expenses", "products", "reports"}

	rolesData := []struct {
		Name        string
		Description string
		CanCreate   bool
		CanRead     bool
		CanUpdate   bool
		CanDelete   bool
	}{
		{Name: "İzle", Description: "Tüm modülleri sadece görüntüleme yetkisi", CanCreate: false, CanRead: true, CanUpdate: false, CanDelete: false},
		{Name: "Oluştur-İzle-Düzenle", Description: "Tüm modülleri görüntüleme, ekleme ve düzenleme yetkisi", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: false},
		{Name: "Oluştur-İzle-Düzenle-Sil", Description: "Tüm modüllerde tam yetki (ekleme, okuma, düzenleme ve silme)", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
	}

	for _, company := range companies {
		for _, rd := range rolesData {
			var role models.Role
			err := SystemDB.Where("company_id = ? AND name = ?", company.ID, rd.Name).First(&role).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				roleID, _ := uuid.NewV7()
				newRole := models.Role{
					ID:          roleID,
					CompanyID:   company.ID,
					Name:        rd.Name,
					Description: rd.Description,
				}
				if err := SystemDB.Create(&newRole).Error; err != nil {
					log.Printf("[migrate] Default role tohumlama hatasi (company: %s, role: %s): %v", company.ID, rd.Name, err)
					continue
				}

				// Create permissions for each module
				var perms []models.RolePermission
				for _, m := range modules {
					permID, _ := uuid.NewV7()
					perms = append(perms, models.RolePermission{
						ID:        permID,
						RoleID:    roleID,
						Module:    m,
						CanCreate: rd.CanCreate,
						CanRead:   rd.CanRead,
						CanUpdate: rd.CanUpdate,
						CanDelete: rd.CanDelete,
					})
				}
				if err := SystemDB.Create(&perms).Error; err != nil {
					log.Printf("[migrate] Default role permissions tohumlama hatasi (company: %s, role: %s): %v", company.ID, rd.Name, err)
				}
			}
		}
	}
}
