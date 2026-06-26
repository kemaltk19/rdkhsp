package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"os"

	"github.com/google/uuid"
)

// MockPaymentProvider, banka API'si hazır olana kadar Charge'ı her zaman
// başarılı döner — Subscribe/Renew akışını gerçek bir provider gibi
// test etmeyi sağlar, gerçek para hareketi olmaz.
type MockPaymentProvider struct{}

func (m *MockPaymentProvider) Charge(ctx context.Context, req ChargeRequest) (*ChargeResult, error) {
	return &ChargeResult{
		Success:     true,
		ProviderRef: "mock_" + uuid.NewString()[:8],
		Message:     "Approved",
	}, nil
}

// VerifyWebhook, gerçek sağlayıcı gelene kadar webhook imzasını
// WEBHOOK_SECRET ile HMAC-SHA256 üzerinden doğrular. Böylece public
// endpoint mock'ta bile yetkisiz isteklere kapalı kalır.
//   - WEBHOOK_SECRET tanımlı değilse: imza doğrulanamaz, istek REDDEDİLİR
//     (güvenli varsayılan; geliştirmede secret set edilmelidir).
//   - signature = hex(HMAC_SHA256(secret, rawBody)) olmalıdır.
func (m *MockPaymentProvider) VerifyWebhook(payload []byte, signature string) (bool, error) {
	secret := os.Getenv("WEBHOOK_SECRET")
	if secret == "" {
		// Secret yoksa doğrulama yapılamaz — güvenli tarafta kal, reddet.
		return false, nil
	}
	if signature == "" {
		return false, nil
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	// Sabit zamanlı karşılaştırma (timing attack'a kapalı).
	return hmac.Equal([]byte(expected), []byte(signature)), nil
}
