package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"radikal-hesap/database"
	"radikal-hesap/models"
	"radikal-hesap/utils"
)

type ExpenseScheduler struct {
	expenseService *ExpenseService
}

func NewExpenseScheduler(es *ExpenseService) *ExpenseScheduler {
	return &ExpenseScheduler{expenseService: es}
}

// Start launches a background routine that runs daily to duplicate recurring expenses
func (s *ExpenseScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		// Run once on startup
		s.processRecurringExpenses(ctx)

		for {
			select {
			case <-ticker.C:
				s.processRecurringExpenses(ctx)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *ExpenseScheduler) processRecurringExpenses(ctx context.Context) {
	log.Println("[scheduler] Tekrarlayan giderler taranıyor...")

	type RecurringExpenseRow struct {
		models.Expense
		CompanyTimezone string `gorm:"column:company_timezone"`
	}

	// Find all expenses marked as recurring that are NOT canceled, joining companies to get their timezone
	var rows []RecurringExpenseRow
	err := database.SystemDB.Table("expenses").
		Select("expenses.*, companies.timezone as company_timezone").
		Joins("left join companies on companies.id = expenses.company_id").
		Where("expenses.is_recurring = ? AND expenses.status != ?", true, "canceled").
		Scan(&rows).Error
	if err != nil {
		log.Printf("[scheduler] Tekrarlayan giderler çekilirken hata oluştu: %v", err)
		return
	}

	for _, row := range rows {
		exp := row.Expense
		companyTimezone := row.CompanyTimezone
		if companyTimezone == "" {
			companyTimezone = "Europe/Istanbul"
		}
		loc := utils.LoadLocation(companyTimezone)
		nowInLoc := time.Now().In(loc)

		// Only duplicate if the current date is after or equal to the expense date (preventing back-generation for old items)
		if nowInLoc.Before(exp.Date) {
			continue
		}

		// We only process if current day equals the transaction day, and no transaction was created for this expense in this month.
		if exp.Date.Day() == nowInLoc.Day() {
			// Check if we already created a recurring copy for this expense in the current month
			var count int64
			err := database.SystemDB.Model(&models.Expense{}).
				Where("category_id = ? AND company_id = ? AND date >= ? AND date < ? AND description LIKE ?", 
					exp.CategoryID, 
					exp.CompanyID, 
					time.Date(nowInLoc.Year(), nowInLoc.Month(), 1, 0, 0, 0, 0, loc), 
					time.Date(nowInLoc.Year(), nowInLoc.Month()+1, 1, 0, 0, 0, 0, loc), 
					fmt.Sprintf("%%(Otomatik Tekrar - Fiş No: %s)%%", exp.ID.String()[:8])).
				Count(&count).Error
			if err != nil {
				log.Printf("[scheduler] Tekrarlayan kontrol sorgusu hatası: %v", err)
				continue
			}

			// Check that we are not trying to copy the parent record itself if it was created in this same month
			if exp.Date.Year() == nowInLoc.Year() && exp.Date.Month() == nowInLoc.Month() {
				continue
			}

			// If no record exists for this month, copy it
			if count == 0 {
				log.Printf("[scheduler] Tekrarlayan gider kopyalanıyor: %s (Company: %s)", exp.Description, exp.CompanyID)
				
				// Build target input payload using the original location
				input := ExpenseInput{
					CategoryID:  exp.CategoryID,
					CariID:      exp.CariID,
					Date:        time.Date(nowInLoc.Year(), nowInLoc.Month(), nowInLoc.Day(), exp.Date.Hour(), exp.Date.Minute(), 0, 0, loc),
					Description: fmt.Sprintf("%s (Otomatik Tekrar - Fiş No: %s)", exp.Description, exp.ID.String()[:8]),
					Amount:      exp.Amount,
					TaxRate:     exp.TaxRate,
					AccountKind: exp.AccountKind,
					AccountID:   exp.AccountID,
					Status:      exp.Status,
					Note:        exp.Note,
					IsRecurring: true, // Keep the chain going
				}

				// Create background context with company tenant info and timezone location
				bgCtx := context.WithValue(context.Background(), "company_id", exp.CompanyID.String())
				bgCtx = context.WithValue(bgCtx, utils.LocationKey, loc)
				creatorID := uuid.Nil
				if exp.CreatedBy != nil {
					creatorID = *exp.CreatedBy
				}

				_, err := s.expenseService.Create(bgCtx, input, creatorID)
				if err != nil {
					log.Printf("[scheduler] Tekrarlayan gider kopyalanamadı: %v", err)
				} else {
					log.Printf("[scheduler] Tekrarlayan gider başarıyla kopyalandı.")
				}
			}
		}
	}
}
