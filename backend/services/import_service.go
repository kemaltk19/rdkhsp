package services

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"

	"radikal-hesap/models"
)

type ImportService struct {
	db *gorm.DB
}

func NewImportService(db *gorm.DB) *ImportService {
	return &ImportService{db: db}
}

// ImportCaris imports cari records from an Excel file, skipping duplicates by Code or TaxNumber
func (s *ImportService) ImportCaris(companyID uuid.UUID, userID uuid.UUID, file *excelize.File) (int, error) {
	sheets := file.GetSheetList()
	if len(sheets) == 0 {
		return 0, errors.New("excel dosyası boş")
	}

	rows, err := file.GetRows(sheets[0])
	if err != nil {
		return 0, err
	}

	importedCount := 0
	// Skip header row
	for i, row := range rows {
		if i == 0 {
			continue
		}
		// Ensure we have enough columns to read safely
		for len(row) < 14 {
			row = append(row, "")
		}

		code := strings.TrimSpace(row[0])
		if code == "" {
			continue
		}

		taxNumber := strings.TrimSpace(row[7])

		// Check duplicate by Code
		var count int64
		s.db.Model(&models.Cari{}).Where("company_id = ? AND code = ?", companyID, code).Count(&count)
		if count > 0 {
			continue // Skip duplicate code
		}

		// Check duplicate by TaxNumber if not empty
		if taxNumber != "" {
			s.db.Model(&models.Cari{}).Where("company_id = ? AND tax_number = ?", companyID, taxNumber).Count(&count)
			if count > 0 {
				continue // Skip duplicate tax number
			}
		}

		cariType := "customer"
		tRaw := strings.ToLower(strings.TrimSpace(row[2]))
		if tRaw == "tedarikçi" || tRaw == "supplier" {
			cariType = "supplier"
		} else if tRaw == "her ikisi" || tRaw == "both" {
			cariType = "both"
		}

		curr := strings.ToUpper(strings.TrimSpace(row[13]))
		if curr == "" {
			curr = "TRY"
		}

		cari := models.Cari{
			ID:          uuid.New(),
			CompanyID:   companyID,
			Code:        code,
			Name:        strings.TrimSpace(row[1]),
			Type:        cariType,
			Group:       strings.TrimSpace(row[3]),
			Title:       strings.TrimSpace(row[4]),
			ContactName: strings.TrimSpace(row[5]),
			TaxOffice:   strings.TrimSpace(row[6]),
			TaxNumber:   taxNumber,
			Email:       strings.TrimSpace(row[8]),
			Phone:       strings.TrimSpace(row[9]),
			Address:     strings.TrimSpace(row[10]),
			City:        strings.TrimSpace(row[11]),
			District:    strings.TrimSpace(row[12]),
			Currency:    curr,
			Country:     "Türkiye",
			IsActive:    true,
			CreatedBy:   &userID,
		}

		if err := s.db.Create(&cari).Error; err == nil {
			importedCount++
		}
	}

	return importedCount, nil
}

// ImportProducts imports product records from an Excel file, skipping duplicates by Code or Barcode
func (s *ImportService) ImportProducts(companyID uuid.UUID, userID uuid.UUID, file *excelize.File) (int, error) {
	sheets := file.GetSheetList()
	if len(sheets) == 0 {
		return 0, errors.New("excel dosyası boş")
	}

	rows, err := file.GetRows(sheets[0])
	if err != nil {
		return 0, err
	}

	importedCount := 0
	// Skip header row
	for i, row := range rows {
		if i == 0 {
			continue
		}
		// Ensure we have enough columns
		for len(row) < 12 {
			row = append(row, "")
		}

		code := strings.TrimSpace(row[0])
		if code == "" {
			continue
		}

		barcode := strings.TrimSpace(row[5])

		// Check duplicate by Code
		var count int64
		s.db.Model(&models.Product{}).Where("company_id = ? AND code = ?", companyID, code).Count(&count)
		if count > 0 {
			continue // Skip duplicate code
		}

		// Check duplicate by Barcode if not empty
		if barcode != "" {
			s.db.Model(&models.Product{}).Where("company_id = ? AND barcode = ?", companyID, barcode).Count(&count)
			if count > 0 {
				continue // Skip duplicate barcode
			}
		}

		pType := "product"
		tRaw := strings.ToLower(strings.TrimSpace(row[3]))
		if tRaw == "hizmet" || tRaw == "service" {
			pType = "service"
		}

		purchPrice, _ := decimal.NewFromString(strings.ReplaceAll(row[7], ",", "."))
		salePrice, _ := decimal.NewFromString(strings.ReplaceAll(row[8], ",", "."))
		taxRate, _ := decimal.NewFromString(strings.ReplaceAll(row[10], ",", "."))

		curr := strings.ToUpper(strings.TrimSpace(row[9]))
		if curr == "" {
			curr = "TRY"
		}

		product := models.Product{
			ID:            uuid.New(),
			CompanyID:     companyID,
			Code:          code,
			Name:          strings.TrimSpace(row[1]),
			Brand:         strings.TrimSpace(row[2]),
			Type:          pType,
			Unit:          strings.TrimSpace(row[4]),
			Barcode:       barcode,
			CustomCodes:   strings.TrimSpace(row[6]),
			PurchasePrice: purchPrice,
			SalePrice:     salePrice,
			Currency:      curr,
			TaxIncluded:   false,
			PurchaseTaxIncluded: false,
			TaxRate:       taxRate,
			PurchaseTaxRate: taxRate,
			Description:   strings.TrimSpace(row[11]),
			TrackStock:    pType == "product",
			IsActive:      true,
			CreatedBy:     &userID,
		}

		if err := s.db.Create(&product).Error; err == nil {
			importedCount++
		}
	}

	return importedCount, nil
}
