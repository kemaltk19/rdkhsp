package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"

	"radikal-hesap/services"
)

type ImportHandler struct {
	importService *services.ImportService
}

func NewImportHandler(importService *services.ImportService) *ImportHandler {
	return &ImportHandler{importService: importService}
}

// ImportCaris handles Excel upload for Cari records
func (h *ImportHandler) ImportCaris(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	userID, _ := c.Get("user_id")

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dosya yüklenemedi"})
		return
	}
	defer file.Close()

	xlsx, err := excelize.OpenReader(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Excel dosyası okunamadı"})
		return
	}
	defer xlsx.Close()

	count, err := h.importService.ImportCaris(companyID.(uuid.UUID), userID.(uuid.UUID), xlsx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("%d adet cari kart başarıyla içe aktarıldı.", count),
		"count":   count,
	})
}

// ImportProducts handles Excel upload for Product records
func (h *ImportHandler) ImportProducts(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	userID, _ := c.Get("user_id")

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dosya yüklenemedi"})
		return
	}
	defer file.Close()

	xlsx, err := excelize.OpenReader(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Excel dosyası okunamadı"})
		return
	}
	defer xlsx.Close()

	count, err := h.importService.ImportProducts(companyID.(uuid.UUID), userID.(uuid.UUID), xlsx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("%d adet stok kartı başarıyla içe aktarıldı.", count),
		"count":   count,
	})
}

// DownloadSampleCari generates and downloads a sample Excel for Caris
func (h *ImportHandler) DownloadSampleCari(c *gin.Context) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Sheet1"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"Cari Kodu", "Firma Adı", "Tip (Müşteri/Tedarikçi)", "Grup", "Resmi Ünvan", "Yetkili Kişi", "Vergi Dairesi", "Vergi No", "E-posta", "Telefon", "Adres", "İl", "İlçe", "Para Birimi"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, header)
	}

	// sample data
	samples := []interface{}{"CAR-001", "Örnek Firma Ltd.", "Müşteri", "Toptancı", "Örnek Firma Ltd. Şti.", "Ahmet Yılmaz", "Kadıköy", "1234567890", "info@ornek.com", "05321234567", "Örnek Mah. Test Sk.", "İstanbul", "Kadıköy", "TRY"}
	for i, val := range samples {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		f.SetCellValue(sheet, cell, val)
	}

	c.Header("Content-Disposition", "attachment; filename=Cari_Iceri_Aktarim_Sablonu.xlsx")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	f.Write(c.Writer)
}

// DownloadSampleProduct generates and downloads a sample Excel for Products
func (h *ImportHandler) DownloadSampleProduct(c *gin.Context) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Sheet1"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"Stok Kodu", "Ürün Adı", "Marka", "Tip (Ürün/Hizmet)", "Birim", "Barkod", "Özel Kodlar", "Alış Fiyatı", "Satış Fiyatı", "Para Birimi", "KDV Oranı (%)", "Açıklama"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, header)
	}

	// sample data
	samples := []interface{}{"STK-001", "Örnek Ürün", "Marka X", "Ürün", "Adet", "8691234567890", "KOD1, KOD2", "100.00", "150.00", "TRY", "20", "Örnek açıklama metni"}
	for i, val := range samples {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		f.SetCellValue(sheet, cell, val)
	}

	c.Header("Content-Disposition", "attachment; filename=Stok_Karti_Iceri_Aktarim_Sablonu.xlsx")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	f.Write(c.Writer)
}
