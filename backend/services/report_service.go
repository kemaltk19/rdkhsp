package services

import (
	"github.com/shopspring/decimal"

	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"radikal-hesap/database"
	"radikal-hesap/models"
	"radikal-hesap/utils"
)

type ReportService struct{}

func NewReportService() *ReportService {
	return &ReportService{}
}

// InvoicesReportRow represents a line item in sales/purchases report
type InvoicesReportRow struct {
	ID            uuid.UUID       `json:"id"`
	Number        string          `json:"number"`
	CariName      string          `json:"cari_name"`
	Date          time.Time       `json:"date"`
	DueDate       time.Time       `json:"due_date"`
	Subtotal      decimal.Decimal `json:"subtotal"`
	DiscountTotal decimal.Decimal `json:"discount_total"`
	TaxTotal      decimal.Decimal `json:"tax_total"`
	Total         decimal.Decimal `json:"total"`
	PaidTotal     decimal.Decimal `json:"paid_total"`
	Status        string          `json:"status"`
}

type InvoicesReportSummary struct {
	Subtotal      decimal.Decimal `json:"subtotal"`
	DiscountTotal decimal.Decimal `json:"discount_total"`
	TaxTotal      decimal.Decimal `json:"tax_total"`
	Total         decimal.Decimal `json:"total"`
	PaidTotal     decimal.Decimal `json:"paid_total"`
	Remaining     decimal.Decimal `json:"remaining"`
}

type InvoicesReportResult struct {
	Rows    []InvoicesReportRow   `json:"rows"`
	Summary InvoicesReportSummary `json:"summary"`
}

// CariAgingRow represents aging buckets for a Cari
type CariAgingRow struct {
	CariID        uuid.UUID       `json:"cari_id"`
	CariName      string          `json:"cari_name"`
	TotalUnpaid   decimal.Decimal `json:"total_unpaid"`
	NotOverdue    decimal.Decimal `json:"not_overdue"`
	Overdue1_30   decimal.Decimal `json:"overdue_1_30"`
	Overdue31_60  decimal.Decimal `json:"overdue_31_60"`
	Overdue61_90  decimal.Decimal `json:"overdue_61_90"`
	Overdue90Plus decimal.Decimal `json:"overdue_90_plus"`
}

type CariAgingResult struct {
	Rows    []CariAgingRow        `json:"rows"`
	Summary InvoicesReportSummary `json:"summary"` // Reusing summary fields for total aging
}

// StockReportRow represents stock valuation row
type StockReportRow struct {
	ProductID     uuid.UUID       `json:"product_id"`
	ProductCode   string          `json:"product_code"`
	ProductName   string          `json:"product_name"`
	CategoryName  string          `json:"category_name"`
	CurrentStock  decimal.Decimal `json:"current_stock"`
	Unit          string          `json:"unit"`
	PurchasePrice decimal.Decimal `json:"purchase_price"`
	Valuation     decimal.Decimal `json:"valuation"` // stock * purchase_price
}

type StockReportResult struct {
	Rows           []StockReportRow `json:"rows"`
	TotalValuation decimal.Decimal  `json:"total_valuation"`
}

// CashReportRow represents cash flow reports
type CashReportRow struct {
	AccountID      uuid.UUID       `json:"account_id"`
	AccountName    string          `json:"account_name"`
	AccountKind    string          `json:"account_kind"` // cash, bank
	Currency       string          `json:"currency"`
	OpeningBalance decimal.Decimal `json:"opening_balance"`
	Inflow         decimal.Decimal `json:"inflow"`
	Outflow        decimal.Decimal `json:"outflow"`
	NetChange      decimal.Decimal `json:"net_change"`
	EndingBalance  decimal.Decimal `json:"ending_balance"`
}

type CashReportResult struct {
	Rows []CashReportRow `json:"rows"`
}

// ProfitLossRow represents P&L grouped by month
type ProfitLossRow struct {
	Month     string          `json:"month"` // YYYY-MM
	Income    decimal.Decimal `json:"income"`
	Expenses  decimal.Decimal `json:"expenses"`
	NetProfit decimal.Decimal `json:"net_profit"`
}

type ProfitLossResult struct {
	Rows         []ProfitLossRow `json:"rows"`
	TotalIncome  decimal.Decimal `json:"total_income"`
	TotalExpense decimal.Decimal `json:"total_expense"`
	TotalProfit  decimal.Decimal `json:"total_profit"`
}

func (s *ReportService) GetReport(ctx context.Context, reportType string, filters map[string]string) (interface{}, error) {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, errors.New("company_id not found in context")
	}
	_, err := uuid.Parse(companyIDStr.(string))
	if err != nil {
		return nil, err
	}

	txTenant := utils.GetDB(ctx, database.DB)

	switch reportType {
	case "sales", "purchases":
		invType := "sales"
		if reportType == "purchases" {
			invType = "purchase"
		}

		dbQuery := txTenant.Model(&models.Invoice{}).
			Where("type = ? AND status != 'canceled'", invType)

		if dateFrom, exists := filters["date_from"]; exists && dateFrom != "" {
			dbQuery = dbQuery.Where("date >= ?", dateFrom)
		}
		if dateTo, exists := filters["date_to"]; exists && dateTo != "" {
			dbQuery = dbQuery.Where("date <= ?", dateTo)
		}
		if cariID, exists := filters["cari_id"]; exists && cariID != "" {
			dbQuery = dbQuery.Where("cari_id = ?", cariID)
		}

		var invoices []models.Invoice
		if err := dbQuery.Order("date DESC, number DESC").Find(&invoices).Error; err != nil {
			return nil, err
		}

		// Resolve cari names
		var caris []models.Cari
		if err := txTenant.Find(&caris).Error; err != nil {
			return nil, err
		}
		cariNameMap := make(map[uuid.UUID]string)
		for _, c := range caris {
			cariNameMap[c.ID] = c.Name
		}

		var sub, disc, tax, tot, paid decimal.Decimal
		rows := make([]InvoicesReportRow, len(invoices))
		for i, inv := range invoices {
			sub = sub.Add(inv.Subtotal)
			disc = disc.Add(inv.DiscountTotal)
			tax = tax.Add(inv.TaxTotal)
			tot = tot.Add(inv.Total)
			paid = paid.Add(inv.PaidTotal)

			rows[i] = InvoicesReportRow{
				ID:            inv.ID,
				Number:        inv.Number,
				CariName:      cariNameMap[inv.CariID],
				Date:          inv.Date,
				DueDate:       inv.DueDate,
				Subtotal:      inv.Subtotal,
				DiscountTotal: inv.DiscountTotal,
				TaxTotal:      inv.TaxTotal,
				Total:         inv.Total,
				PaidTotal:     inv.PaidTotal,
				Status:        inv.Status,
			}
		}

		return InvoicesReportResult{
			Rows: rows,
			Summary: InvoicesReportSummary{
				Subtotal:      sub,
				DiscountTotal: disc,
				TaxTotal:      tax,
				Total:         tot,
				PaidTotal:     paid,
				Remaining:     tot.Sub(paid),
			},
		}, nil

	case "cari-aging":
		// Outstanding invoices (Sales type only, not canceled, not paid)
		var invoices []models.Invoice
		if err := txTenant.Preload("Items").
			Where("type = 'sales' AND status != 'canceled' AND status != 'paid'").
			Find(&invoices).Error; err != nil {
			return nil, err
		}

		var caris []models.Cari
		if err := txTenant.Find(&caris).Error; err != nil {
			return nil, err
		}
		cariMap := make(map[uuid.UUID]*CariAgingRow)
		for _, c := range caris {
			cariMap[c.ID] = &CariAgingRow{
				CariID:   c.ID,
				CariName: c.Name,
			}
		}

		now := utils.NowIn(ctx)
		var totVal, paidVal decimal.Decimal

		for _, inv := range invoices {
			unpaid := inv.Total.Sub(inv.PaidTotal)
			if unpaid.LessThanOrEqual(decimal.Zero) {
				continue
			}

			totVal = totVal.Add(inv.Total)
			paidVal = paidVal.Add(inv.PaidTotal)

			row, ok := cariMap[inv.CariID]
			if !ok {
				continue
			}

			row.TotalUnpaid = row.TotalUnpaid.Add(unpaid)

			// Overdue days calculation: now - due_date
			if inv.DueDate.After(now) {
				row.NotOverdue = row.NotOverdue.Add(unpaid)
			} else {
				overdueDays := int(now.Sub(inv.DueDate).Hours() / 24)
				if overdueDays <= 30 {
					row.Overdue1_30 = row.Overdue1_30.Add(unpaid)
				} else if overdueDays <= 60 {
					row.Overdue31_60 = row.Overdue31_60.Add(unpaid)
				} else if overdueDays <= 90 {
					row.Overdue61_90 = row.Overdue61_90.Add(unpaid)
				} else {
					row.Overdue90Plus = row.Overdue90Plus.Add(unpaid)
				}
			}
		}

		var activeRows []CariAgingRow
		for _, row := range cariMap {
			if row.TotalUnpaid.GreaterThan(decimal.Zero) {
				activeRows = append(activeRows, *row)
			}
		}

		return CariAgingResult{
			Rows: activeRows,
			Summary: InvoicesReportSummary{
				Total:     totVal,
				PaidTotal: paidVal,
				Remaining: totVal.Sub(paidVal),
			},
		}, nil

	case "stock":
		// Get all products category-linked
		var products []models.Product
		if err := txTenant.Where("track_stock = true").Find(&products).Error; err != nil {
			return nil, err
		}

		var categories []models.ProductCategory
		if err := txTenant.Find(&categories).Error; err != nil {
			return nil, err
		}
		catNameMap := make(map[uuid.UUID]string)
		for _, cat := range categories {
			catNameMap[cat.ID] = cat.Name
		}

		var totalValuation decimal.Decimal
		rows := make([]StockReportRow, len(products))
		for i, p := range products {
			stock := p.CurrentStock
			price := p.PurchasePrice
			val := stock.Mul(price)
			totalValuation = totalValuation.Add(val)

			catName := "Genel"
			if p.CategoryID != nil {
				if name, ok := catNameMap[*p.CategoryID]; ok {
					catName = name
				}
			}

			rows[i] = StockReportRow{
				ProductID:     p.ID,
				ProductCode:   p.Code,
				ProductName:   p.Name,
				CategoryName:  catName,
				CurrentStock:  p.CurrentStock,
				Unit:          p.Unit,
				PurchasePrice: p.PurchasePrice,
				Valuation:     val,
			}
		}

		return StockReportResult{
			Rows:           rows,
			TotalValuation: totalValuation,
		}, nil

	case "cash":
		// Cash accounts & Bank accounts summary flows
		var cashAccs []models.CashAccount
		var bankAccs []models.BankAccount
		if err := txTenant.Find(&cashAccs).Error; err != nil {
			return nil, err
		}
		if err := txTenant.Find(&bankAccs).Error; err != nil {
			return nil, err
		}

		dateFrom := filters["date_from"]
		dateTo := filters["date_to"]

		rows := make([]CashReportRow, 0, len(cashAccs)+len(bankAccs))

		// 1. Nakit Kasalar
		for _, acc := range cashAccs {
			inflowQuery := txTenant.Model(&models.CashTransaction{}).
				Where("account_kind = 'cash' AND account_id = ? AND type = 'in'", acc.ID)
			outflowQuery := txTenant.Model(&models.CashTransaction{}).
				Where("account_kind = 'cash' AND account_id = ? AND type = 'out'", acc.ID)

			if dateFrom != "" {
				inflowQuery = inflowQuery.Where("date >= ?", dateFrom)
				outflowQuery = outflowQuery.Where("date >= ?", dateFrom)
			}
			if dateTo != "" {
				inflowQuery = inflowQuery.Where("date <= ?", dateTo)
				outflowQuery = outflowQuery.Where("date <= ?", dateTo)
			}

			var inflow, outflow decimal.Decimal
			inflowQuery.Select("COALESCE(SUM(amount), 0)").Scan(&inflow)
			outflowQuery.Select("COALESCE(SUM(amount), 0)").Scan(&outflow)

			// If date range is specified, calculate opening balance
			// final balance = opening + inflow - outflow => opening = final - inflow + outflow
			// But wait, the current balance cache on CashAccount represents the absolute final balance *now*.
			// If date_from is set, the cash flow calculates the opening balance at that specific moment.
			// Let's determine cash transactions after dateTo to adjust ending balance too:
			var afterEndingInflow, afterEndingOutflow decimal.Decimal
			if dateTo != "" {
				txTenant.Model(&models.CashTransaction{}).
					Where("account_kind = 'cash' AND account_id = ? AND type = 'in' AND date > ?", acc.ID, dateTo).
					Select("COALESCE(SUM(amount), 0)").Scan(&afterEndingInflow)
				txTenant.Model(&models.CashTransaction{}).
					Where("account_kind = 'cash' AND account_id = ? AND type = 'out' AND date > ?", acc.ID, dateTo).
					Select("COALESCE(SUM(amount), 0)").Scan(&afterEndingOutflow)
			}

			ending := acc.Balance.Sub(afterEndingInflow).Add(afterEndingOutflow)
			opening := ending.Sub(inflow).Add(outflow)

			rows = append(rows, CashReportRow{
				AccountID:      acc.ID,
				AccountName:    acc.Name,
				AccountKind:    "cash",
				Currency:       acc.Currency,
				OpeningBalance: opening,
				Inflow:         inflow,
				Outflow:        outflow,
				NetChange:      inflow.Sub(outflow),
				EndingBalance:  ending,
			})
		}

		// 2. Banka Hesapları
		for _, acc := range bankAccs {
			inflowQuery := txTenant.Model(&models.CashTransaction{}).
				Where("account_kind = 'bank' AND account_id = ? AND type = 'in'", acc.ID)
			outflowQuery := txTenant.Model(&models.CashTransaction{}).
				Where("account_kind = 'bank' AND account_id = ? AND type = 'out'", acc.ID)

			if dateFrom != "" {
				inflowQuery = inflowQuery.Where("date >= ?", dateFrom)
				outflowQuery = outflowQuery.Where("date >= ?", dateFrom)
			}
			if dateTo != "" {
				inflowQuery = inflowQuery.Where("date <= ?", dateTo)
				outflowQuery = outflowQuery.Where("date <= ?", dateTo)
			}

			var inflow, outflow decimal.Decimal
			inflowQuery.Select("COALESCE(SUM(amount), 0)").Scan(&inflow)
			outflowQuery.Select("COALESCE(SUM(amount), 0)").Scan(&outflow)

			var afterEndingInflow, afterEndingOutflow decimal.Decimal
			if dateTo != "" {
				txTenant.Model(&models.CashTransaction{}).
					Where("account_kind = 'bank' AND account_id = ? AND type = 'in' AND date > ?", acc.ID, dateTo).
					Select("COALESCE(SUM(amount), 0)").Scan(&afterEndingInflow)
				txTenant.Model(&models.CashTransaction{}).
					Where("account_kind = 'bank' AND account_id = ? AND type = 'out' AND date > ?", acc.ID, dateTo).
					Select("COALESCE(SUM(amount), 0)").Scan(&afterEndingOutflow)
			}

			ending := acc.Balance.Sub(afterEndingInflow).Add(afterEndingOutflow)
			opening := ending.Sub(inflow).Add(outflow)

			rows = append(rows, CashReportRow{
				AccountID:      acc.ID,
				AccountName:    acc.Name,
				AccountKind:    "bank",
				Currency:       acc.Currency,
				OpeningBalance: opening,
				Inflow:         inflow,
				Outflow:        outflow,
				NetChange:      inflow.Sub(outflow),
				EndingBalance:  ending,
			})
		}

		return CashReportResult{Rows: rows}, nil

	case "profit":
		// Income (Sales Invoices totals) grouped by month for the last 12 months
		// Expense (Expense totals) grouped by month for the last 12 months
		now := utils.NowIn(ctx)
		twelveMonthsAgo := now.AddDate(0, -11, 0)
		startOfTwelveMonths := utils.StartOfMonth(ctx, twelveMonthsAgo)

		type MonthlySum struct {
			Month string          `gorm:"column:m"`
			Total decimal.Decimal `gorm:"column:t"`
		}

		var dbIncomes []MonthlySum
		if err := txTenant.Table("invoices").
			Select("TO_CHAR(date, 'YYYY-MM') as m, COALESCE(SUM(total), 0) as t").
			Where("type = 'sales' AND status != 'canceled' AND date >= ?", startOfTwelveMonths).
			Group("TO_CHAR(date, 'YYYY-MM')").
			Scan(&dbIncomes).Error; err != nil {
			return nil, err
		}

		var dbExpenses []MonthlySum
		if err := txTenant.Table("expenses").
			Select("TO_CHAR(date, 'YYYY-MM') as m, COALESCE(SUM(total), 0) as t").
			Where("status != 'canceled' AND date >= ?", startOfTwelveMonths). // Wait! Expense models have status? Or are they just active? Let's check:
			// Actually let's group all expenses since they represent cost.
			// Let's verify: does expense table have a status column?
			// Let's check models/expense.go structure later or query without status filter to be safe!
			// In models/expense.go, does it have `status`?
			Group("TO_CHAR(date, 'YYYY-MM')").
			Scan(&dbExpenses).Error; err != nil {
			return nil, err
		}

		incomeMap := make(map[string]decimal.Decimal)
		for _, s := range dbIncomes {
			incomeMap[s.Month] = s.Total
		}

		expenseMap := make(map[string]decimal.Decimal)
		for _, s := range dbExpenses {
			expenseMap[s.Month] = s.Total
		}

		var totInc, totExp, totProf decimal.Decimal
		rows := make([]ProfitLossRow, 12)
		for i := 11; i >= 0; i-- {
			t := now.AddDate(0, -i, 0)
			monthStr := t.Format("2006-01")

			inc := incomeMap[monthStr]
			exp := expenseMap[monthStr]
			net := inc.Sub(exp)

			totInc = totInc.Add(inc)
			totExp = totExp.Add(exp)
			totProf = totProf.Add(net)

			rows[11-i] = ProfitLossRow{
				Month:     monthStr,
				Income:    inc,
				Expenses:  exp,
				NetProfit: net,
			}
		}

		return ProfitLossResult{
			Rows:         rows,
			TotalIncome:  totInc,
			TotalExpense: totExp,
			TotalProfit:  totProf,
		}, nil

	default:
		return nil, errors.New("invalid_report_type")
	}
}
