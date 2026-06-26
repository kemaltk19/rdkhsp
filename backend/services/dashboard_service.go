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

type DashboardService struct{}

func NewDashboardService() *DashboardService {
	return &DashboardService{}
}

type RecentCariTx struct {
	ID          uuid.UUID       `json:"id"`
	CariName    string          `json:"cari_name"`
	Date        time.Time       `json:"date"`
	Type        string          `json:"type"` // debit, credit
	SourceType  string          `json:"source_type"`
	Description string          `json:"description"`
	Amount      decimal.Decimal `json:"amount"`
}

type RecentCashTx struct {
	ID          uuid.UUID       `json:"id"`
	AccountName string          `json:"account_name"`
	AccountKind string          `json:"account_kind"` // cash, bank
	Date        time.Time       `json:"date"`
	Type        string          `json:"type"` // in, out
	Amount      decimal.Decimal `json:"amount"`
	Description string          `json:"description"`
}

type RecentExpenseTx struct {
	ID           uuid.UUID       `json:"id"`
	CategoryName string          `json:"category_name"`
	Date         time.Time       `json:"date"`
	Amount       decimal.Decimal `json:"amount"`
	Currency     string          `json:"currency"`
	Description  string          `json:"description"`
}

type ChartDataPoint struct {
	Month string          `json:"month"` // e.g. "2026-06"
	Total decimal.Decimal `json:"total"`
}

type CurrencyAmount struct {
	Currency string          `json:"currency" gorm:"column:currency"`
	Amount   decimal.Decimal `json:"amount" gorm:"column:amount"`
}

type DashboardStats struct {
	Ciro          []CurrencyAmount  `json:"ciro"`
	ToCollect     []CurrencyAmount  `json:"to_collect"`
	CashBankTotal []CurrencyAmount  `json:"cash_bank_total"`
	OverdueTotal  []CurrencyAmount  `json:"overdue_total"`
	RecentCariTx   []RecentCariTx     `json:"recent_cari_tx"`
	RecentCashTx   []RecentCashTx     `json:"recent_cash_tx"`
	RecentExpenses []RecentExpenseTx  `json:"recent_expenses"`
	ChartData      []ChartDataPoint   `json:"chart_data"`
	ChartSeries    []ChartSeriesData  `json:"chart_series"` // Multi-currency series
}

type ChartSeriesData struct {
	Currency string           `json:"currency"`
	Data     []ChartDataPoint `json:"data"`
}

type DBChartPoint struct {
	M string  `gorm:"column:m"`
	C string  `gorm:"column:currency"`
	T float64 `gorm:"column:t"`
}

func (s *DashboardService) GetStats(ctx context.Context) (*DashboardStats, error) {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, errors.New("company_id not found in context")
	}
	_, err := uuid.Parse(companyIDStr.(string))
	if err != nil {
		return nil, err
	}

	txTenant := utils.GetDB(ctx, database.DB)
	now := utils.NowIn(ctx)

	// 1. Current Month Ciro (Sales invoices status != canceled)
	var ciroVals []CurrencyAmount
	startOfMonth := utils.StartOfMonth(ctx, now)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)
	if err := txTenant.Model(&models.Invoice{}).
		Where("type = 'sales' AND status != 'canceled' AND status != 'draft' AND date >= ? AND date <= ?", startOfMonth, endOfMonth).
		Select("currency, COALESCE(SUM(total), 0) as amount").
		Group("currency").Scan(&ciroVals).Error; err != nil {
		return nil, err
	}

	// 2. Pending Collections (Sales invoices total - paid_total, status != paid/canceled)
	var toCollectVals []CurrencyAmount
	if err := txTenant.Model(&models.Invoice{}).
		Where("type = 'sales' AND status != 'canceled' AND status != 'draft' AND status != 'paid'").
		Select("currency, COALESCE(SUM(total - paid_total), 0) as amount").
		Group("currency").Scan(&toCollectVals).Error; err != nil {
		return nil, err
	}

	// 3. Cash & Bank Balances Total
	var cashVals []CurrencyAmount
	if err := txTenant.Model(&models.CashAccount{}).Select("currency, COALESCE(SUM(balance), 0) as amount").Group("currency").Scan(&cashVals).Error; err != nil {
		return nil, err
	}
	var bankVals []CurrencyAmount
	if err := txTenant.Model(&models.BankAccount{}).Select("currency, COALESCE(SUM(balance), 0) as amount").Group("currency").Scan(&bankVals).Error; err != nil {
		return nil, err
	}
	
	cashBankMap := make(map[string]decimal.Decimal)
	for _, v := range cashVals {
		cashBankMap[v.Currency] = cashBankMap[v.Currency].Add(v.Amount)
	}
	for _, v := range bankVals {
		cashBankMap[v.Currency] = cashBankMap[v.Currency].Add(v.Amount)
	}
	var cashBankTotal []CurrencyAmount
	for cur, val := range cashBankMap {
		cashBankTotal = append(cashBankTotal, CurrencyAmount{Currency: cur, Amount: val})
	}

	// 4. Overdue Invoices Total (due_date < now, type = sales, status != paid/canceled)
	var overdueVals []CurrencyAmount
	if err := txTenant.Model(&models.Invoice{}).
		Where("type = 'sales' AND status != 'canceled' AND status != 'draft' AND status != 'paid' AND due_date < ?", now).
		Select("currency, COALESCE(SUM(total - paid_total), 0) as amount").
		Group("currency").Scan(&overdueVals).Error; err != nil {
		return nil, err
	}

	// 5. Recent Cari Transactions
	var recentCariTxs []RecentCariTx
	if err := txTenant.Table("cari_transactions").
		Select("cari_transactions.id, caris.name as cari_name, cari_transactions.date, cari_transactions.type, cari_transactions.source_type, cari_transactions.description, cari_transactions.amount").
		Joins("left join caris on caris.id = cari_transactions.cari_id").
		Order("cari_transactions.date DESC, cari_transactions.created_at DESC").
		Limit(5).
		Scan(&recentCariTxs).Error; err != nil {
		return nil, err
	}
	if recentCariTxs == nil {
		recentCariTxs = []RecentCariTx{}
	}

	// 5.5 Recent Expenses
	var recentExpenses []RecentExpenseTx
	if err := txTenant.Table("expenses").
		Select("expenses.id, expense_categories.name as category_name, expenses.date, expenses.amount, expenses.currency, expenses.description").
		Joins("left join expense_categories on expense_categories.id = expenses.category_id").
		Order("expenses.date DESC, expenses.created_at DESC").
		Limit(5).
		Scan(&recentExpenses).Error; err != nil {
		return nil, err
	}
	if recentExpenses == nil {
		recentExpenses = []RecentExpenseTx{}
	}

	// 6. Recent Cash Transactions (resolve cash/bank account names)
	var rawCashTxs []models.CashTransaction
	if err := txTenant.Order("date DESC, created_at DESC").Limit(5).Find(&rawCashTxs).Error; err != nil {
		return nil, err
	}

	var cashAccs []models.CashAccount
	var bankAccs []models.BankAccount
	if err := txTenant.Find(&cashAccs).Error; err != nil {
		return nil, err
	}
	if err := txTenant.Find(&bankAccs).Error; err != nil {
		return nil, err
	}

	accountNameMap := make(map[uuid.UUID]string)
	for _, acc := range cashAccs {
		accountNameMap[acc.ID] = acc.Name
	}
	for _, acc := range bankAccs {
		accountNameMap[acc.ID] = acc.Name
	}

	recentCashTxs := make([]RecentCashTx, len(rawCashTxs))
	for i, tx := range rawCashTxs {
		name := accountNameMap[tx.AccountID]
		if name == "" {
			if tx.AccountKind == "cash" {
				name = "Kasa Hesabı"
			} else {
				name = "Banka Hesabı"
			}
		}
		recentCashTxs[i] = RecentCashTx{
			ID:          tx.ID,
			AccountName: name,
			AccountKind: tx.AccountKind,
			Date:        tx.Date,
			Type:        tx.Type,
			Amount:      tx.Amount,
			Description: tx.Description,
		}
	}

	// 7. Last 6 Months Chart Data
	chartData := make([]ChartDataPoint, 6)
	for i := 5; i >= 0; i-- {
		t := now.AddDate(0, -i, 0)
		chartData[5-i] = ChartDataPoint{
			Month: t.Format("2006-01"),
			Total: decimal.Zero,
		}
	}

	// Start of chart data (5 months ago, 1st day, 00:00:00)
	fiveMonthsAgo := now.AddDate(0, -5, 0)
	startOfSixMonths := utils.StartOfMonth(ctx, fiveMonthsAgo)

	var dbPoints []DBChartPoint
	if err := txTenant.Table("invoices").
		Select("TO_CHAR(date, 'YYYY-MM') as m, currency, COALESCE(SUM(total), 0) as t").
		Where("type = 'sales' AND status != 'canceled' AND status != 'draft' AND date >= ?", startOfSixMonths).
		Group("TO_CHAR(date, 'YYYY-MM'), currency").
		Scan(&dbPoints).Error; err != nil {
		return nil, err
	}

	// Organize points by currency
	currencyPoints := make(map[string]map[string]float64)
	for _, p := range dbPoints {
		if _, ok := currencyPoints[p.C]; !ok {
			currencyPoints[p.C] = make(map[string]float64)
		}
		currencyPoints[p.C][p.M] = p.T
	}

	var chartSeries []ChartSeriesData
	for curr, monthsMap := range currencyPoints {
		seriesData := make([]ChartDataPoint, 6)
		for i := 0; i < 6; i++ {
			monthStr := chartData[i].Month
			val := monthsMap[monthStr]
			seriesData[i] = ChartDataPoint{
				Month: monthStr,
				Total: decimal.NewFromFloat(val),
			}
		}
		chartSeries = append(chartSeries, ChartSeriesData{
			Currency: curr,
			Data:     seriesData,
		})
	}

	// For legacy support or fallback
	dbPointsMap := make(map[string]float64)
	for _, p := range dbPoints {
		dbPointsMap[p.M] += p.T // sums everything for the old ChartData
	}
	for i := 0; i < 6; i++ {
		if val, ok := dbPointsMap[chartData[i].Month]; ok {
			chartData[i].Total = decimal.NewFromFloat(val)
		}
	}

	return &DashboardStats{
		Ciro:          ciroVals,
		ToCollect:     toCollectVals,
		CashBankTotal: cashBankTotal,
		OverdueTotal:  overdueVals,
		RecentCariTx:   recentCariTxs,
		RecentCashTx:   recentCashTxs,
		RecentExpenses: recentExpenses,
		ChartData:      chartData,
		ChartSeries:    chartSeries,
	}, nil
}
