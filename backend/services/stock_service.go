package services

import (
	"fmt"
	"github.com/shopspring/decimal"

	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"radikal-hesap/database"
	"radikal-hesap/models"
	"radikal-hesap/utils"
)

var (
	ErrProductNotFound         = errors.New("product_not_found")
	ErrProductCategoryNotFound = errors.New("product_category_not_found")
	ErrWarehouseNotFound       = errors.New("warehouse_not_found")
	ErrDuplicateProductCode    = errors.New("duplicate_product_code")
	ErrProductInUse            = errors.New("product_in_use_cannot_delete")
	ErrCategoryInUse           = errors.New("category_in_use_cannot_delete")
	ErrWarehouseInUse          = errors.New("warehouse_in_use_cannot_delete")
)

type StockService struct{}

func NewStockService() *StockService {
	return &StockService{}
}

type ProductInput struct {
	Code          string          `json:"code" binding:"max=100"`
	Name          string          `json:"name" binding:"required,max=255"`
	Brand         string          `json:"brand" binding:"max=100"`
	Type          string          `json:"type" binding:"required,oneof=product service"`
	Unit          string          `json:"unit" binding:"max=50"`
	Barcode       string          `json:"barcode" binding:"max=100"`
	CustomCodes   string          `json:"custom_codes" binding:"max=255"`
	SerialNumbers string          `json:"serial_numbers" binding:"max=2000"`
	PurchasePrice decimal.Decimal `json:"purchase_price"`
	SalePrice     decimal.Decimal `json:"sale_price"`
	Currency      string          `json:"currency" binding:"max=10"`
	TaxIncluded   bool            `json:"tax_included"`
	PurchaseTaxIncluded bool      `json:"purchase_tax_included"`
	TaxRate       decimal.Decimal `json:"tax_rate"`
	PurchaseTaxRate decimal.Decimal `json:"purchase_tax_rate"`
	Description   string          `json:"description" binding:"max=2000"`
	TrackStock    bool            `json:"track_stock"`
	MinStock      decimal.Decimal `json:"min_stock"`
	InitialStock  decimal.Decimal `json:"initial_stock"`
	CategoryID    *uuid.UUID      `json:"category_id"`
	IsActive      bool            `json:"is_active"`
}

func (s *StockService) CreateProduct(ctx context.Context, in ProductInput, createdBy uuid.UUID) (*models.Product, error) {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, errors.New("company_id not found in context")
	}
	companyID, _ := uuid.Parse(companyIDStr.(string))

	var product models.Product

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		// 1. Verify category if provided
		if in.CategoryID != nil {
			var cat models.ProductCategory
			if err := txTenant.First(&cat, "id = ?", *in.CategoryID).Error; err != nil {
				return ErrProductCategoryNotFound
			}
		}

		// 2. Generate code if empty
		code := strings.TrimSpace(in.Code)
		if in.Code == "" {
			// Çakışmaya dayanıklı üretim: sayaç ile gerçek veri arasında bir
			// tutarsızlık varsa (örn. geçmişte sayaç artmadan kod oluşmuşsa),
			// boş bir kod bulana kadar sayacı atlatarak ilerle.
			for attempt := 0; attempt < 1000; attempt++ {
				generated, err := utils.GenerateNumberWithSetting(txTenant, companyID, "product", "product_prefix", "SKU")
				if err != nil {
					return err
				}
				var count int64
				if err := txTenant.Model(&models.Product{}).Where("company_id = ? AND code = ?", companyID, generated).Count(&count).Error; err != nil {
					return err
				}
				if count == 0 {
					code = generated
					break
				}
			}
			if code == "" {
				return errors.New("benzersiz ürün kodu üretilemedi")
			}
		} else {
			// Verify code uniqueness
			var count int64
			if err := txTenant.Model(&models.Product{}).Where("company_id = ? AND code = ?", companyID, code).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return ErrDuplicateProductCode
			}
		}

		productID, _ := uuid.NewV7()
		product = models.Product{
			ID:            productID,
			CompanyID:     companyID,
			Code:          code,
			Name:          in.Name,
			Brand:         in.Brand,
			Type:          in.Type,
			Unit:          in.Unit,
			Barcode:       in.Barcode,
			CustomCodes:   in.CustomCodes,
			SerialNumbers: in.SerialNumbers,
			PurchasePrice: in.PurchasePrice,
			AverageCost:   in.PurchasePrice,
			SalePrice:     in.SalePrice,
			Currency:      in.Currency,
			TaxIncluded:   in.TaxIncluded,
			PurchaseTaxIncluded: in.PurchaseTaxIncluded,
			TaxRate:       in.TaxRate,
			PurchaseTaxRate: in.PurchaseTaxRate,
			Description:   in.Description,
			TrackStock:    in.TrackStock,
			CurrentStock:  decimal.Zero,
			MinStock:      in.MinStock,
			CategoryID:    in.CategoryID,
			IsActive:      in.IsActive,
			CreatedBy:     &createdBy,
		}

		if err := txTenant.Create(&product).Error; err != nil {
			return err
		}

		if in.InitialStock.GreaterThan(decimal.Zero) && in.TrackStock && in.Type == "product" {
			var wh models.Warehouse
			if err := txTenant.First(&wh, "company_id = ? AND is_default = ?", companyID, true).Error; err != nil {
				txTenant.First(&wh, "company_id = ?", companyID)
			}
			if wh.ID != uuid.Nil {
				smID, _ := uuid.NewV7()
				sm := models.StockMovement{
					ID:           smID,
					CompanyID:    companyID,
					ProductID:    product.ID,
					WarehouseID:  wh.ID,
					Date:         time.Now(),
					Type:         "in",
					SourceType:   "manual",
					Quantity:     in.InitialStock,
					UnitCost:     in.PurchasePrice,
					BalanceAfter: in.InitialStock,
					Note:         "Açılış Stoğu",
					CreatedBy:    &createdBy,
				}
				if err := txTenant.Create(&sm).Error; err != nil {
					return err
				}
				product.CurrentStock = in.InitialStock
				if err := txTenant.Save(&product).Error; err != nil {
					return err
				}
			}
		}

		if err := WriteAuditLog(ctx, txTenant, "product", product.ID, "create", createdBy, product.Name); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (s *StockService) UpdateProduct(ctx context.Context, id uuid.UUID, in ProductInput, userID uuid.UUID) (*models.Product, error) {
	txTenant := utils.GetDB(ctx, database.DB)

	var product models.Product
	if err := txTenant.First(&product, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	// Verify category if changed and provided
	if in.CategoryID != nil && (product.CategoryID == nil || *product.CategoryID != *in.CategoryID) {
		var cat models.ProductCategory
		if err := txTenant.First(&cat, "id = ?", *in.CategoryID).Error; err != nil {
			return nil, ErrProductCategoryNotFound
		}
	}

	// Verify Code uniqueness if changed
	code := strings.TrimSpace(in.Code)
	if code != "" && code != product.Code {
		var count int64
		if err := txTenant.Model(&models.Product{}).Where("company_id = ? AND code = ? AND id != ?", product.CompanyID, code, id).Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, ErrDuplicateProductCode
		}
		product.Code = code
	}

	product.Name = in.Name
	product.Brand = in.Brand
	product.Type = in.Type
	product.Unit = in.Unit
	product.Barcode = in.Barcode
	product.CustomCodes = in.CustomCodes
	product.SerialNumbers = in.SerialNumbers
	product.PurchasePrice = in.PurchasePrice
	product.SalePrice = in.SalePrice
	product.Currency = in.Currency
	product.TaxIncluded = in.TaxIncluded
	product.PurchaseTaxIncluded = in.PurchaseTaxIncluded
	product.TaxRate = in.TaxRate
	product.PurchaseTaxRate = in.PurchaseTaxRate
	product.Description = in.Description
	product.TrackStock = in.TrackStock
	product.MinStock = in.MinStock
	product.CategoryID = in.CategoryID
	product.IsActive = in.IsActive

	if err := txTenant.Save(&product).Error; err != nil {
		return nil, err
	}

	if err := WriteAuditLog(ctx, txTenant, "product", product.ID, "update", userID, product.Name); err != nil {
		return nil, err
	}

	return &product, nil
}

func (s *StockService) DeleteProduct(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	txTenant := utils.GetDB(ctx, database.DB)

	// Check if in use in stock movements or invoice items
	var count int64
	if err := txTenant.Model(&models.StockMovement{}).Where("product_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrProductInUse
	}

	if err := txTenant.Model(&models.InvoiceItem{}).Where("product_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrProductInUse
	}

	var product models.Product
	if err := txTenant.First(&product, "id = ?", id).Error; err == nil {
		if err := WriteAuditLog(ctx, txTenant, "product", product.ID, "delete", userID, product.Name); err != nil {
			return err
		}
	}

	return txTenant.Delete(&models.Product{}, "id = ?", id).Error
}

func (s *StockService) GetProductByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	txTenant := utils.GetDB(ctx, database.DB)
	var product models.Product
	if err := txTenant.Preload("Category").First(&product, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return &product, nil
}

func (s *StockService) ListProducts(ctx context.Context, page, limit int, query, sort string, filters map[string]string) ([]models.Product, int64, error) {
	txTenant := utils.GetDB(ctx, database.DB)

	var list []models.Product
	var total int64

	dbQuery := txTenant.Model(&models.Product{}).Preload("Category")

	if query != "" {
		q := "%" + query + "%"
		dbQuery = dbQuery.Where("name ILIKE ? OR code ILIKE ? OR barcode ILIKE ? OR custom_codes ILIKE ? OR brand ILIKE ?", q, q, q, q, q)
	}

	if typeFilter, exists := filters["type"]; exists && typeFilter != "" {
		dbQuery = dbQuery.Where("type = ?", typeFilter)
	}

	if catFilter, exists := filters["category_id"]; exists && catFilter != "" {
		dbQuery = dbQuery.Where("category_id = ?", catFilter)
	}

	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if sort != "" {
		dbQuery = dbQuery.Order(sort)
	} else {
		dbQuery = dbQuery.Order("code ASC, created_at DESC")
	}

	offset := (page - 1) * limit
	if err := dbQuery.Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// ----------------------------------------------------
// Stock Movement Management
// ----------------------------------------------------

type StockMovementInput struct {
	ProductID   uuid.UUID       `json:"product_id" binding:"required"`
	WarehouseID uuid.UUID       `json:"warehouse_id"` // optional, will default to company's default
	Date        time.Time       `json:"date" binding:"required"`
	Type        string          `json:"type" binding:"required,oneof=in out adjustment"`
	Quantity    decimal.Decimal `json:"quantity" binding:"required"`
	UnitCost    decimal.Decimal `json:"unit_cost"`
	Note        string          `json:"note"`
}

func (s *StockService) ManualAdjustment(ctx context.Context, in StockMovementInput, createdBy uuid.UUID) (*models.StockMovement, error) {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, errors.New("company_id not found in context")
	}
	companyID, _ := uuid.Parse(companyIDStr.(string))

	var movement models.StockMovement

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		// 1. Verify Product
		var product models.Product
		if err := txTenant.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, "id = ?", in.ProductID).Error; err != nil {
			return ErrProductNotFound
		}

		// 2. Determine Warehouse
		warehouseID := in.WarehouseID
		if warehouseID == uuid.Nil {
			defWhID, err := s.GetOrCreateDefaultWarehouse(txTenant, companyID, createdBy)
			if err != nil {
				return err
			}
			warehouseID = defWhID
		} else {
			var wh models.Warehouse
			if err := txTenant.First(&wh, "id = ?", warehouseID).Error; err != nil {
				return ErrWarehouseNotFound
			}
		}

		// 3. Compute Stock Balance and Average Cost
		var balanceAfter decimal.Decimal
		newAverageCost := product.AverageCost
		qty := in.Quantity

		if in.Type == "in" {
			balanceAfter = product.CurrentStock.Add(qty)
			if balanceAfter.GreaterThan(decimal.Zero) {
				// (Eski Stok * Eski Maliyet) + (Giren Stok * Giren Maliyet) / Yeni Toplam Stok
				totalOldValue := product.CurrentStock.Mul(product.AverageCost)
				totalNewValue := qty.Mul(in.UnitCost)
				newAverageCost = totalOldValue.Add(totalNewValue).DivRound(balanceAfter, 4)
			}
		} else if in.Type == "out" {
			balanceAfter = product.CurrentStock.Sub(qty)
			if balanceAfter.IsNegative() {
				return errors.New("insufficient_stock_for_adjustment")
			}
		} else { // adjustment
			// For adjustments: quantity represents the new absolute stock value
			balanceAfter = qty
			qty = balanceAfter.Sub(product.CurrentStock)
			if qty.GreaterThanOrEqual(decimal.Zero) {
				in.Type = "in"
				if balanceAfter.GreaterThan(decimal.Zero) {
					totalOldValue := product.CurrentStock.Mul(product.AverageCost)
					totalNewValue := qty.Mul(in.UnitCost)
					newAverageCost = totalOldValue.Add(totalNewValue).DivRound(balanceAfter, 4)
				}
			} else {
				in.Type = "out"
				qty = qty.Neg()
			}
		}

		movementID, _ := uuid.NewV7()
		movement = models.StockMovement{
			ID:           movementID,
			CompanyID:    companyID,
			ProductID:    in.ProductID,
			WarehouseID:  warehouseID,
			Date:         in.Date,
			Type:         in.Type,
			SourceType:   "manual",
			Quantity:     qty,
			UnitCost:     in.UnitCost,
			BalanceAfter: balanceAfter,
			Note:         in.Note,
			CreatedBy:    &createdBy,
		}

		if err := txTenant.Create(&movement).Error; err != nil {
			return err
		}

		summaryStr := fmt.Sprintf("%s - %s %s - %s", product.Name, in.Type, qty.StringFixed(2), product.Unit)
		if err := WriteAuditLog(ctx, txTenant, "stock_movement", movement.ID, "create", createdBy, summaryStr); err != nil {
			return err
		}

		// Update product cached stock and average cost
		return txTenant.Model(&product).Updates(map[string]interface{}{
			"current_stock": balanceAfter,
			"average_cost":  newAverageCost,
		}).Error
	})

	if err != nil {
		return nil, err
	}
	return &movement, nil
}

func (s *StockService) ListMovements(ctx context.Context, productID uuid.UUID) ([]models.StockMovement, error) {
	txTenant := utils.GetDB(ctx, database.DB)
	var list []models.StockMovement
	err := txTenant.Preload("Warehouse").Where("product_id = ?", productID).Order("date DESC, created_at DESC").Find(&list).Error
	return list, err
}

// ----------------------------------------------------
// Warehouse Service
// ----------------------------------------------------

func (s *StockService) CreateWarehouse(ctx context.Context, name, address string, isDefault bool, createdBy uuid.UUID) (*models.Warehouse, error) {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, errors.New("company_id not found")
	}
	companyID, _ := uuid.Parse(companyIDStr.(string))

	txTenant := utils.GetDB(ctx, database.DB)

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		txTenantTx := utils.GetDB(ctx, tx)

		if isDefault {
			// Remove other defaults
			txTenantTx.Model(&models.Warehouse{}).Where("company_id = ?", companyID).Update("is_default", false)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	whID, _ := uuid.NewV7()
	wh := models.Warehouse{
		ID:        whID,
		CompanyID: companyID,
		Name:      name,
		Address:   address,
		IsDefault: isDefault,
		CreatedBy: &createdBy,
	}

	if err := txTenant.Create(&wh).Error; err != nil {
		return nil, err
	}

	if err := WriteAuditLog(ctx, txTenant, "warehouse", wh.ID, "create", createdBy, wh.Name); err != nil {
		return nil, err
	}

	return &wh, nil
}

func (s *StockService) UpdateWarehouse(ctx context.Context, id uuid.UUID, name, address string, isDefault bool, userID uuid.UUID) (*models.Warehouse, error) {
	txTenant := utils.GetDB(ctx, database.DB)

	var count int64
	if err := txTenant.Model(&models.StockMovement{}).Where("warehouse_id = ?", id).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrWarehouseInUse
	}

	var wh models.Warehouse
	if err := txTenant.First(&wh, "id = ?", id).Error; err != nil {
		return nil, ErrWarehouseNotFound
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		txTenantTx := utils.GetDB(ctx, tx)
		if isDefault && !wh.IsDefault {
			txTenantTx.Model(&models.Warehouse{}).Where("company_id = ?", wh.CompanyID).Update("is_default", false)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	wh.Name = name
	wh.Address = address
	wh.IsDefault = isDefault

	if err := txTenant.Save(&wh).Error; err != nil {
		return nil, err
	}

	if err := WriteAuditLog(ctx, txTenant, "warehouse", wh.ID, "update", userID, wh.Name); err != nil {
		return nil, err
	}

	return &wh, nil
}

func (s *StockService) DeleteWarehouse(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	txTenant := utils.GetDB(ctx, database.DB)
	var count int64
	if err := txTenant.Model(&models.StockMovement{}).Where("warehouse_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrWarehouseInUse
	}

	var wh models.Warehouse
	if err := txTenant.First(&wh, "id = ?", id).Error; err == nil {
		if err := WriteAuditLog(ctx, txTenant, "warehouse", wh.ID, "delete", userID, wh.Name); err != nil {
			return err
		}
	}

	return txTenant.Delete(&models.Warehouse{}, "id = ?", id).Error
}

func (s *StockService) ListWarehouses(ctx context.Context) ([]models.Warehouse, error) {
	txTenant := utils.GetDB(ctx, database.DB)
	var list []models.Warehouse
	err := txTenant.Order("name ASC").Find(&list).Error
	return list, err
}

func (s *StockService) GetOrCreateDefaultWarehouse(txTenant *gorm.DB, companyID uuid.UUID, createdBy uuid.UUID) (uuid.UUID, error) {
	var wh models.Warehouse
	err := txTenant.Where("company_id = ? AND is_default = ?", companyID, true).First(&wh).Error
	if err == nil {
		return wh.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return uuid.Nil, err
	}

	// Try first available warehouse
	err = txTenant.Where("company_id = ?", companyID).First(&wh).Error
	if err == nil {
		return wh.ID, nil
	}

	// Provision default warehouse
	whID, _ := uuid.NewV7()
	wh = models.Warehouse{
		ID:        whID,
		CompanyID: companyID,
		Name:      "Merkez Depo",
		Address:   "Merkez Deposu",
		IsDefault: true,
		CreatedBy: &createdBy,
	}
	if err := txTenant.Create(&wh).Error; err != nil {
		return uuid.Nil, err
	}
	return wh.ID, nil
}

// ----------------------------------------------------
// Product Category Service
// ----------------------------------------------------

func (s *StockService) CreateCategory(ctx context.Context, name string, defaultKDVRate decimal.Decimal, createdBy uuid.UUID) (*models.ProductCategory, error) {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, errors.New("company_id not found")
	}
	companyID, _ := uuid.Parse(companyIDStr.(string))

	txTenant := utils.GetDB(ctx, database.DB)

	catID, _ := uuid.NewV7()
	cat := models.ProductCategory{
		ID:             catID,
		CompanyID:      companyID,
		Name:           name,
		DefaultKDVRate: defaultKDVRate,
		CreatedBy:      &createdBy,
	}

	if err := txTenant.Create(&cat).Error; err != nil {
		return nil, err
	}

	if err := WriteAuditLog(ctx, txTenant, "product_category", cat.ID, "create", createdBy, cat.Name); err != nil {
		return nil, err
	}

	return &cat, nil
}

func (s *StockService) UpdateCategory(ctx context.Context, id uuid.UUID, name string, defaultKDVRate decimal.Decimal, userID uuid.UUID) (*models.ProductCategory, error) {
	txTenant := utils.GetDB(ctx, database.DB)

	var count int64
	if err := txTenant.Model(&models.Product{}).Where("category_id = ?", id).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrCategoryInUse
	}

	var cat models.ProductCategory
	if err := txTenant.First(&cat, "id = ?", id).Error; err != nil {
		return nil, ErrProductCategoryNotFound
	}
	cat.Name = name
	cat.DefaultKDVRate = defaultKDVRate

	if err := txTenant.Save(&cat).Error; err != nil {
		return nil, err
	}

	if err := WriteAuditLog(ctx, txTenant, "product_category", cat.ID, "update", userID, cat.Name); err != nil {
		return nil, err
	}

	return &cat, nil
}

func (s *StockService) DeleteCategory(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	txTenant := utils.GetDB(ctx, database.DB)
	var count int64
	if err := txTenant.Model(&models.Product{}).Where("category_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrCategoryInUse
	}

	var cat models.ProductCategory
	if err := txTenant.First(&cat, "id = ?", id).Error; err == nil {
		if err := WriteAuditLog(ctx, txTenant, "product_category", cat.ID, "delete", userID, cat.Name); err != nil {
			return err
		}
	}

	return txTenant.Delete(&models.ProductCategory{}, "id = ?", id).Error
}

func (s *StockService) ListCategories(ctx context.Context) ([]models.ProductCategory, error) {
	txTenant := utils.GetDB(ctx, database.DB)
	var list []models.ProductCategory
	err := txTenant.Order("name ASC").Find(&list).Error
	return list, err
}

const DefaultCriticalThreshold = int64(5)

type CriticalStockItem struct {
	ID           uuid.UUID       `json:"id"`
	Code         string          `json:"code"`
	Name         string          `json:"name"`
	Unit         string          `json:"unit"`
	CurrentStock decimal.Decimal `json:"current_stock"`
	Threshold    decimal.Decimal `json:"threshold"`
}

func (s *StockService) GetCriticalStockProducts(ctx context.Context) ([]CriticalStockItem, error) {
	txTenant := utils.GetDB(ctx, database.DB)

	var products []models.Product
	err := txTenant.
		Where(`
            track_stock = true
            AND type = 'product'
            AND is_active = true
            AND current_stock <= CASE WHEN min_stock > 0 THEN min_stock ELSE 5 END
        `).
		Find(&products).Error
	if err != nil {
		return nil, err
	}

	result := make([]CriticalStockItem, 0, len(products))
	for _, p := range products {
		threshold := p.MinStock
		if threshold.IsZero() || threshold.IsNegative() {
			threshold = decimal.NewFromInt(DefaultCriticalThreshold)
		}
		result = append(result, CriticalStockItem{
			ID:           p.ID,
			Code:         p.Code,
			Name:         p.Name,
			Unit:         p.Unit,
			CurrentStock: p.CurrentStock,
			Threshold:    threshold,
		})
	}
	return result, nil
}

func (s *StockService) GetNextCode(ctx context.Context) (map[string]interface{}, error) {
	txTenant := utils.GetDB(ctx, database.DB)
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, errors.New("company_id not found in context")
	}
	companyID, err := uuid.Parse(companyIDStr.(string))
	if err != nil {
		return nil, err
	}

	code, err := utils.PreviewNumberWithSetting(txTenant, companyID, "product", "product_prefix", "SKU")
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{"next_code": code}, nil
}
