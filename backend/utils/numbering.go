package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GenerateNumber increments and returns a unique sequential number for the company and key.
// It uses `SELECT FOR UPDATE` on `number_sequences` table within the current transaction to prevent race conditions.
func GenerateNumber(tx *gorm.DB, companyID uuid.UUID, key string, prefix string) (string, error) {
	// Guard insert to ensure row exists
	var count int64
	err := tx.Table("number_sequences").Where("company_id = ? AND key = ?", companyID, key).Count(&count).Error
	if err != nil {
		return "", err
	}

	if count == 0 {
		insertQuery := "INSERT INTO number_sequences (company_id, key, last_no) VALUES (?, ?, 0) ON CONFLICT DO NOTHING"
		if err := tx.Exec(insertQuery, companyID, key).Error; err != nil {
			return "", err
		}
	}

	var seq struct {
		LastNo int
	}

	// Query with SELECT FOR UPDATE row lock
	selectQuery := "SELECT last_no FROM number_sequences WHERE company_id = ? AND key = ? FOR UPDATE"
	err = tx.Raw(selectQuery, companyID, key).Scan(&seq).Error
	if err != nil {
		return "", err
	}

	nextNo := seq.LastNo + 1

	// Update sequence counter
	updateQuery := "UPDATE number_sequences SET last_no = ? WHERE company_id = ? AND key = ?"
	if err := tx.Exec(updateQuery, nextNo, companyID, key).Error; err != nil {
		return "", err
	}

	return SmartFormat(prefix, nextNo), nil
}

var trailingDigitRegex = regexp.MustCompile(`(\d+)$`)

// SmartFormat extracts trailing digits from prefix, increments them, and formats them back.
// If the prefix does not end with a digit, it falls back to prefix-%05d
func SmartFormat(prefix string, count int) string {
	if !trailingDigitRegex.MatchString(prefix) {
		separator := "-"
		if strings.HasSuffix(prefix, "-") || strings.HasSuffix(prefix, ".") {
			separator = ""
		}
		return fmt.Sprintf("%s%s%05d", prefix, separator, count)
	}

	// count=1 özel durumu: prefix'in sonundaki rakamı 1 artır
	// (0'dan döngüye girmemek için prefix'i olduğu gibi iade etmek yanlıştı —
	// farklı sequence key'ler aynı prefix'i paylaştığında çakışıyordu).
	matches := trailingDigitRegex.FindStringSubmatchIndex(prefix)
	if len(matches) < 4 {
		return prefix
	}

	start, end := matches[2], matches[3]
	numStr := prefix[start:end]

	num, _ := strconv.Atoi(numStr)
	newNum := num + (count - 1)

	format := fmt.Sprintf("%%0%dd", len(numStr))
	newNumStr := fmt.Sprintf(format, newNum)

	return prefix[:start] + newNumStr
}

// GenerateNumberWithSetting reads the prefix from the `settings` table dynamically.
// If the setting is missing or empty, it falls back to defaultPrefix.
func GenerateNumberWithSetting(tx *gorm.DB, companyID uuid.UUID, key string, settingKey string, defaultPrefix string) (string, error) {
	prefix := defaultPrefix
	var val string
	// Query the settings table. We use tx.Table to avoid cyclic dependencies with models package.
	err := tx.Table("settings").Select("value").Where("company_id = ? AND key = ?", companyID, settingKey).Scan(&val).Error
	if err == nil && val != "" {
		prefix = val
	}
	return GenerateNumber(tx, companyID, key, prefix)
}

// PreviewNumber returns the next sequential number without incrementing it.
func PreviewNumber(tx *gorm.DB, companyID uuid.UUID, key string, prefix string) (string, error) {
	var count int64
	err := tx.Table("number_sequences").Where("company_id = ? AND key = ?", companyID, key).Count(&count).Error
	if err != nil {
		return "", err
	}

	if count == 0 {
		return SmartFormat(prefix, 1), nil
	}

	var seq struct {
		LastNo int
	}

	err = tx.Raw("SELECT last_no FROM number_sequences WHERE company_id = ? AND key = ?", companyID, key).Scan(&seq).Error
	if err != nil {
		return "", err
	}

	return SmartFormat(prefix, seq.LastNo+1), nil
}

// PreviewNumberWithSetting reads the prefix from the `settings` table dynamically and previews the next number.
func PreviewNumberWithSetting(tx *gorm.DB, companyID uuid.UUID, key string, settingKey string, defaultPrefix string) (string, error) {
	prefix := defaultPrefix
	var val string
	err := tx.Table("settings").Select("value").Where("company_id = ? AND key = ?", companyID, settingKey).Scan(&val).Error
	if err == nil && val != "" {
		prefix = val
	}
	return PreviewNumber(tx, companyID, key, prefix)
}
