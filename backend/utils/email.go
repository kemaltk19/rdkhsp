package utils

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"log"
	"mime"
	"mime/quotedprintable"
	"net/smtp"
	"strings"
	"time"
)

// messageID üretir: Gmail gibi alıcılar Message-ID başlığı olmayan mailleri
// (Date'i de) standart dışı sayıp sessizce reddedebilir/droplayabilir.
func messageID(domain string) string {
	var buf [16]byte
	_, _ = rand.Read(buf[:])
	if domain == "" {
		domain = "radikalhesap.com"
	}
	return fmt.Sprintf("<%d.%s@%s>", time.Now().UnixNano(), hex.EncodeToString(buf[:]), domain)
}

// senderDomain, From adresinin domain kısmını döndürür (Message-ID için).
func senderDomain(from string) string {
	if i := strings.LastIndex(from, "@"); i >= 0 && i+1 < len(from) {
		return from[i+1:]
	}
	return ""
}

// Email is a renderable message.
type Email struct {
	To      string
	Subject string
	HTML    string
	// FromName, doluysa SMTP config'indeki sabit görünen adı (FromName) ezer.
	// Gönderen ADRESİ değişmez (kimlik doğrulama ona bağlı), sadece alıcının
	// gelen kutusunda gördüğü ad değişir. Örn: belge maillerinde firma adı.
	FromName string
}

// SMTPConfig is an explicit SMTP configuration (used for DB-managed settings).
type SMTPConfig struct {
	Host, Port, User, Pass, From, FromName string
}

// Mailer sends emails.
type Mailer interface {
	Send(e Email) error
}

// DefaultMailer is the env/log fallback mailer set by InitMailer.
var DefaultMailer Mailer = &logMailer{}

// InitMailer selects an env-based SMTP mailer when host is set, otherwise a log mailer.
func InitMailer(host, port, user, pass, from string) {
	if strings.TrimSpace(host) == "" {
		log.Println("[mail] SMTP_HOST bos -> log mailer (mailler konsola yazilacak)")
		DefaultMailer = &logMailer{}
		return
	}
	DefaultMailer = &smtpMailer{cfg: SMTPConfig{Host: host, Port: port, User: user, Pass: pass, From: from}}
	log.Printf("[mail] SMTP mailer (env) aktif (%s:%s)", host, port)
}

// SendEmail sends via the default (env/log) mailer. Prefer services.SendEmail for DB-managed settings.
func SendEmail(e Email) error {
	if DefaultMailer == nil {
		DefaultMailer = &logMailer{}
	}
	return DefaultMailer.Send(e)
}

type logMailer struct{}

func (l *logMailer) Send(e Email) error {
	log.Printf("[mail:log] To=%s | Subject=%s\n%s", e.To, e.Subject, e.HTML)
	return nil
}

type smtpMailer struct{ cfg SMTPConfig }

func (s *smtpMailer) Send(e Email) error { return SendSMTP(s.cfg, e) }

// SendSMTP sends one HTML email using the given SMTP config.
// Port 465 uses implicit TLS (the connection is encrypted from the first byte);
// net/smtp.SendMail cannot do this, it only speaks STARTTLS (port 587/25).
func SendSMTP(cfg SMTPConfig, e Email) error {
	if strings.TrimSpace(cfg.Host) == "" {
		log.Printf("[mail:log] (host yok) To=%s | %s", e.To, e.Subject)
		return nil
	}
	port := cfg.Port
	if port == "" {
		port = "587"
	}
	// Görünen ad önceliği: maile özel FromName (örn. firma adı) > SMTP varsayılanı.
	fromName := cfg.FromName
	if e.FromName != "" {
		fromName = e.FromName
	}
	header := cfg.From
	if fromName != "" {
		header = fmt.Sprintf("%s <%s>", mime.QEncoding.Encode("utf-8", fromName), cfg.From)
	}
	subject := mime.QEncoding.Encode("utf-8", e.Subject)

	// Gövdeyi quoted-printable ile kodla: Türkçe/emoji (non-ASCII) içeren HTML'i
	// "8bit" olarak göndermek bazı alıcılarda (özellikle Gmail) teslimatı bozuyordu.
	var qpBody strings.Builder
	qp := quotedprintable.NewWriter(&qpBody)
	_, _ = qp.Write([]byte(e.HTML))
	_ = qp.Close()

	var b strings.Builder
	b.WriteString("From: " + header + "\r\n")
	b.WriteString("To: " + e.To + "\r\n")
	b.WriteString("Subject: " + subject + "\r\n")
	// Date ve Message-ID zorunlu: bunlar olmadan Gmail maili sessizce
	// reddediyor/droplyor (sunucu RCPT'te kabul etse bile kutuya düşmüyor).
	b.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\r\n")
	b.WriteString("Message-ID: " + messageID(senderDomain(cfg.From)) + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	b.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
	b.WriteString("\r\n")
	b.WriteString(qpBody.String())

	var auth smtp.Auth
	if cfg.User != "" {
		auth = smtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host)
	}
	addr := cfg.Host + ":" + port

	if port == "465" {
		return sendSMTPImplicitTLS(addr, cfg.Host, auth, cfg.From, e.To, []byte(b.String()))
	}

	if err := smtp.SendMail(addr, auth, cfg.From, []string{e.To}, []byte(b.String())); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	return nil
}

// sendSMTPImplicitTLS connects over TLS from the first byte (port 465 style),
// then drives the standard SMTP command sequence on top of it.
func sendSMTPImplicitTLS(addr, host string, auth smtp.Auth, from, to string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return fmt.Errorf("smtp tls dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt to: %w", err)
	}
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := wc.Write(msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("smtp data close: %w", err)
	}
	return client.Quit()
}

// ---- Locale-aware templates (locale: "tr" | "en") ----

func loc(locale string) string {
	if locale == "en" {
		return "en"
	}
	return "tr"
}

func wrapHTML(title, body string) string {
	return fmt.Sprintf(`<!DOCTYPE html><html><body style="margin:0;padding:0;background:#eef1f6;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;">`+
		`<div style="max-width:520px;margin:0 auto;padding:32px 16px;">`+
		`<div style="text-align:center;margin-bottom:20px;">`+
		`<span style="font-size:18px;font-weight:800;color:#2563eb;letter-spacing:-0.5px;">Radikal Hesap</span></div>`+
		`<div style="background:#fff;border-radius:14px;padding:36px;box-shadow:0 4px 16px rgba(15,23,42,0.06);">`+
		`<h2 style="color:#0f172a;margin:0 0 16px;font-size:21px;">%s</h2>`+
		`<div style="color:#334155;font-size:15px;line-height:1.6;">%s</div>`+
		`</div>`+
		`<div style="text-align:center;margin-top:28px;color:#94a3b8;font-size:12px;line-height:1.6;">`+
		`<p style="margin:0 0 4px;">Bu e-posta <a href="https://radikalhesap.com" style="color:#64748b;text-decoration:none;font-weight:600;">radikalhesap.com</a> üzerinden otomatik olarak gönderildi.</p>`+
		`<p style="margin:0;">&copy; %d Radikal Hesap. Tüm hakları saklıdır.</p>`+
		`</div></div></body></html>`, title, body, time.Now().Year())
}

func button(url, label string) string {
	return fmt.Sprintf(`<p style="text-align:center;margin:28px 0 16px;"><a href="%s" style="display:inline-block;background:#2563eb;color:#fff;text-decoration:none;padding:13px 28px;border-radius:8px;font-weight:600;font-size:14px;">%s</a></p>`+
		`<p style="color:#94a3b8;font-size:12px;word-break:break-all;text-align:center;">%s</p>`, url, label, url)
}

// VerificationEmail builds the email-verification message.
func VerificationEmail(locale, link string) Email {
	if loc(locale) == "en" {
		return Email{Subject: "Verify your email", HTML: wrapHTML("Verify your email",
			"<p>Welcome! Please confirm your email address to activate your account.</p>"+button(link, "Verify email"))}
	}
	return Email{Subject: "E-postanı doğrula", HTML: wrapHTML("E-postanı doğrula",
		"<p>Hoş geldin! Hesabını etkinleştirmek için e-posta adresini doğrula.</p>"+button(link, "E-postayı doğrula"))}
}

// PasswordResetEmail builds the password-reset message.
func PasswordResetEmail(locale, link string) Email {
	if loc(locale) == "en" {
		return Email{Subject: "Reset your password", HTML: wrapHTML("Reset your password",
			"<p>We received a request to reset your password. This link expires in 1 hour. If you didn't request it, ignore this email.</p>"+button(link, "Reset password"))}
	}
	return Email{Subject: "Şifreni sıfırla", HTML: wrapHTML("Şifreni sıfırla",
		"<p>Şifre sıfırlama talebi aldık. Bu bağlantı 1 saat içinde geçerliliğini yitirir. Sen talep etmediysen yok say.</p>"+button(link, "Şifreyi sıfırla"))}
}

// WelcomeEmail builds a post-registration welcome message.
func WelcomeEmail(locale, name string) Email {
	if loc(locale) == "en" {
		return Email{Subject: "Welcome to Radikal Hesap", HTML: wrapHTML("Welcome",
			fmt.Sprintf("<p>Hi %s, your account is ready. Enjoy your trial!</p>", name))}
	}
	return Email{Subject: "Radikal Hesap'a hoş geldin", HTML: wrapHTML("Hoş geldin",
		fmt.Sprintf("<p>Merhaba %s, hesabın hazır. Deneme süreni iyi değerlendir!</p>", name))}
}

// CompanySummary is the sender-company profile shown in document notification
// emails (invoice/quote/etc.), so the recipient sees who it's from at a glance.
type CompanySummary struct {
	Name    string
	Phone   string
	Email   string
	Address string
}

func (c CompanySummary) blockHTML(locale string) string {
	var lines []string
	if c.Phone != "" {
		lines = append(lines, fmt.Sprintf(`<span>📞 %s</span>`, c.Phone))
	}
	if c.Email != "" {
		lines = append(lines, fmt.Sprintf(`<span>✉️ %s</span>`, c.Email))
	}
	if c.Address != "" {
		lines = append(lines, fmt.Sprintf(`<span>📍 %s</span>`, c.Address))
	}
	contact := strings.Join(lines, " &nbsp;|&nbsp; ")
	return fmt.Sprintf(`<div style="background:#f8fafc;border-radius:10px;padding:16px 18px;margin-bottom:22px;border:1px solid #e2e8f0;">`+
		`<div style="font-weight:700;color:#0f172a;margin-bottom:6px;font-size:15px;">%s</div>`+
		`<div style="font-size:12.5px;color:#64748b;line-height:1.5;">%s</div></div>`, c.Name, contact)
}

// documentSubject builds the "[Belge Etiketi] [Belge No]" subject line shared by
// all outbound document notifications (invoice, quote, and future modules).
func documentSubject(docLabel, docNumber string) string {
	return fmt.Sprintf("%s %s", docLabel, docNumber)
}

// DocumentLine is one row of the items table embedded in invoice/quote emails.
type DocumentLine struct {
	Description string
	Quantity    string
	Unit        string
	UnitPrice   string
	Total       string
}

// DocumentParty is the customer/recipient block shown on the document card.
type DocumentParty struct {
	Name    string
	Address string
}

// DocumentSummary carries everything needed to render the invoice/quote-style
// HTML card embedded in the notification email (header, party, items, totals).
type DocumentSummary struct {
	Number       string
	Date         string
	DueDate      string // vade/geçerlilik tarihi
	Customer     DocumentParty
	Lines        []DocumentLine
	Subtotal     string
	DiscountText string // boşsa satır gösterilmez
	TaxText      string
	Total        string
	Currency     string
}

func (d DocumentSummary) itemsHTML() string {
	var rows strings.Builder
	for _, l := range d.Lines {
		rows.WriteString(fmt.Sprintf(
			`<tr>`+
				`<td style="padding:10px 12px;border-bottom:1px solid #eef1f6;color:#0f172a;font-size:13px;">%s</td>`+
				`<td style="padding:10px 12px;border-bottom:1px solid #eef1f6;color:#475569;font-size:13px;text-align:center;white-space:nowrap;">%s %s</td>`+
				`<td style="padding:10px 12px;border-bottom:1px solid #eef1f6;color:#475569;font-size:13px;text-align:right;white-space:nowrap;">%s</td>`+
				`<td style="padding:10px 12px;border-bottom:1px solid #eef1f6;color:#0f172a;font-size:13px;text-align:right;white-space:nowrap;font-weight:600;">%s</td>`+
				`</tr>`, l.Description, l.Quantity, l.Unit, l.UnitPrice, l.Total))
	}

	totalsRows := fmt.Sprintf(
		`<tr><td style="padding:4px 0;color:#64748b;font-size:13px;">Ara Toplam</td><td style="padding:4px 0;text-align:right;color:#0f172a;font-size:13px;">%s</td></tr>`,
		d.Subtotal)
	if d.DiscountText != "" {
		totalsRows += fmt.Sprintf(
			`<tr><td style="padding:4px 0;color:#64748b;font-size:13px;">İndirim</td><td style="padding:4px 0;text-align:right;color:#0f172a;font-size:13px;">%s</td></tr>`,
			d.DiscountText)
	}
	totalsRows += fmt.Sprintf(
		`<tr><td style="padding:4px 0;color:#64748b;font-size:13px;">KDV</td><td style="padding:4px 0;text-align:right;color:#0f172a;font-size:13px;">%s</td></tr>`,
		d.TaxText)
	totalsRows += fmt.Sprintf(
		`<tr><td style="padding:10px 0 0;color:#0f172a;font-size:15px;font-weight:800;border-top:1px solid #e2e8f0;">Genel Toplam</td>`+
			`<td style="padding:10px 0 0;text-align:right;color:#2563eb;font-size:17px;font-weight:800;border-top:1px solid #e2e8f0;">%s</td></tr>`,
		d.Total)

	return fmt.Sprintf(
		`<div style="border:1px solid #e2e8f0;border-radius:12px;overflow:hidden;margin:20px 0;">`+
			`<div style="background:#f8fafc;padding:14px 16px;display:flex;justify-content:space-between;flex-wrap:wrap;gap:8px;border-bottom:1px solid #e2e8f0;">`+
			`<div><div style="font-size:11px;color:#94a3b8;text-transform:uppercase;letter-spacing:.04em;">SAYIN</div>`+
			`<div style="font-size:14px;font-weight:700;color:#0f172a;">%s</div></div>`+
			`<div style="text-align:right;"><div style="font-size:11px;color:#94a3b8;text-transform:uppercase;letter-spacing:.04em;">TARİH / VADE</div>`+
			`<div style="font-size:13px;color:#334155;">%s — %s</div></div></div>`+
			`<table style="width:100%%;border-collapse:collapse;">`+
			`<thead><tr style="background:#fff;">`+
			`<th style="text-align:left;padding:10px 12px;font-size:11px;color:#94a3b8;text-transform:uppercase;letter-spacing:.04em;border-bottom:1px solid #e2e8f0;">Açıklama</th>`+
			`<th style="text-align:center;padding:10px 12px;font-size:11px;color:#94a3b8;text-transform:uppercase;letter-spacing:.04em;border-bottom:1px solid #e2e8f0;">Miktar</th>`+
			`<th style="text-align:right;padding:10px 12px;font-size:11px;color:#94a3b8;text-transform:uppercase;letter-spacing:.04em;border-bottom:1px solid #e2e8f0;">Birim Fiyat</th>`+
			`<th style="text-align:right;padding:10px 12px;font-size:11px;color:#94a3b8;text-transform:uppercase;letter-spacing:.04em;border-bottom:1px solid #e2e8f0;">Toplam</th>`+
			`</tr></thead><tbody>%s</tbody></table>`+
			`<div style="padding:14px 16px;background:#fafbfc;"><table style="width:100%%;max-width:280px;margin-left:auto;border-collapse:collapse;">%s</table></div>`+
			`</div>`,
		d.Customer.Name, d.Date, d.DueDate, rows.String(), totalsRows)
}

// InvoiceEmail builds the invoice delivery message: sender company summary,
// a PDF-style invoice card (items/tax/totals), and the public view/pay link.
func InvoiceEmail(locale string, company CompanySummary, invoiceNumber, amount string, doc DocumentSummary, link string) Email {
	if loc(locale) == "en" {
		subject := documentSubject("Invoice", invoiceNumber)
		body := company.blockHTML(locale) +
			fmt.Sprintf("<p>You have received a new invoice (<b>%s</b>) for <b>%s</b>.</p>", invoiceNumber, amount) +
			doc.itemsHTML() +
			button(link, "View & Pay Invoice")
		return Email{Subject: subject, HTML: wrapHTML("Invoice Summary", body), FromName: company.Name}
	}
	subject := documentSubject("Fatura", invoiceNumber)
	body := company.blockHTML(locale) +
		fmt.Sprintf("<p>Tarafınıza <b>%s</b> tutarında <b>%s</b> numaralı yeni bir fatura kesilmiştir.</p>", amount, invoiceNumber) +
		doc.itemsHTML() +
		button(link, "Faturayı Görüntüle ve Öde")
	return Email{Subject: subject, HTML: wrapHTML("Fatura Özeti", body), FromName: company.Name}
}

// PurchaseInvoiceEmail, alış faturalarında tedarikçiye gönderilen TEYİT mailidir.
// Alış faturasını tedarikçi bize kesmiştir; ona "ödeyin" demek (ödeme linki)
// mantıksız olur. Bu yüzden ödeme butonu YOKTUR — sadece faturanın tarafımızca
// sisteme kaydedildiğini/alındığını bildirir ve belge özetini gösterir.
func PurchaseInvoiceEmail(locale string, company CompanySummary, invoiceNumber, amount string, doc DocumentSummary) Email {
	if loc(locale) == "en" {
		subject := documentSubject("Invoice received", invoiceNumber)
		body := company.blockHTML(locale) +
			fmt.Sprintf("<p>Your invoice <b>%s</b> for <b>%s</b> has been received and recorded in our system.</p>", invoiceNumber, amount) +
			doc.itemsHTML()
		return Email{Subject: subject, HTML: wrapHTML("Invoice Received", body), FromName: company.Name}
	}
	subject := documentSubject("Faturanız alındı", invoiceNumber)
	body := company.blockHTML(locale) +
		fmt.Sprintf("<p>Tarafınızca kesilen <b>%s</b> tutarındaki <b>%s</b> numaralı fatura sistemimize kaydedilmiştir. Bilginize sunarız.</p>", amount, invoiceNumber) +
		doc.itemsHTML()
	return Email{Subject: subject, HTML: wrapHTML("Faturanız Alındı", body), FromName: company.Name}
}

// QuoteEmail builds the quote delivery message: sender company summary,
// a PDF-style quote card (items/tax/totals), and the public view/respond link.
func QuoteEmail(locale string, company CompanySummary, number, amount string, doc DocumentSummary, link string) Email {
	if loc(locale) == "en" {
		subject := documentSubject("Quote", number)
		body := company.blockHTML(locale) +
			fmt.Sprintf("<p>You have received a new quote (<b>%s</b>) for <b>%s</b>.</p>", number, amount) +
			doc.itemsHTML() +
			button(link, "View & Respond")
		return Email{Subject: subject, HTML: wrapHTML("Quote Summary", body), FromName: company.Name}
	}
	subject := documentSubject("Teklif", number)
	body := company.blockHTML(locale) +
		fmt.Sprintf("<p>Tarafınıza <b>%s</b> tutarında <b>%s</b> numaralı yeni bir teklif iletilmiştir.</p>", amount, number) +
		doc.itemsHTML() +
		button(link, "Teklifi Görüntüle ve Yanıtla")
	return Email{Subject: subject, HTML: wrapHTML("Teklif Özeti", body), FromName: company.Name}
}
