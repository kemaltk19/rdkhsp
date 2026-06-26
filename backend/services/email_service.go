package services

import (
	"log"

	"radikal-hesap/database"
	"radikal-hesap/models"
	"radikal-hesap/utils"
)

// resolveSMTP returns the DB-configured (superadmin) SMTP if enabled,
// otherwise (zero, false) to fall back to env/log mailer.
func resolveSMTP() (utils.SMTPConfig, bool) {
	var s models.EmailSetting
	if err := database.SystemDB.First(&s).Error; err != nil {
		log.Printf("[mail] DB SMTP ayarı okunamadı (%v) -> log mailer'a düşülüyor", err)
		return utils.SMTPConfig{}, false
	}
	if !s.Enabled || s.Host == "" {
		log.Printf("[mail] DB SMTP devre dışı (enabled=%v, host_bos=%v) -> log mailer'a düşülüyor", s.Enabled, s.Host == "")
		return utils.SMTPConfig{}, false
	}
	pass, _ := utils.Decrypt(s.PasswordEnc)
	return utils.SMTPConfig{
		Host:     s.Host,
		Port:     s.Port,
		User:     s.Username,
		Pass:     pass,
		From:     s.FromEmail,
		FromName: s.FromName,
	}, true
}

// SendEmail sends using the superadmin's DB SMTP settings when configured,
// otherwise the env/log fallback. Auth flows should call this.
func SendEmail(e utils.Email) error {
	if cfg, ok := resolveSMTP(); ok {
		return utils.SendSMTP(cfg, e)
	}
	return utils.SendEmail(e)
}
