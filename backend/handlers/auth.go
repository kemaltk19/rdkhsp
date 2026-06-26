package handlers

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"radikal-hesap/services"
	"radikal-hesap/utils"
)

type AuthHandler struct {
	svc          *services.AuthService
	cookieSecure bool
	sameSite     http.SameSite
}

func NewAuthHandler(svc *services.AuthService) *AuthHandler {
	secure, _ := strconv.ParseBool(os.Getenv("COOKIE_SECURE"))
	return &AuthHandler{
		svc:          svc,
		cookieSecure: secure,
		sameSite:     parseSameSite(),
	}
}

// parseSameSite: COOKIE_SAMESITE = lax (varsayilan) | none | strict.
// Cross-domain (frontend/API farkli domain) deploy'da "none" + COOKIE_SECURE=true gerekir.
func parseSameSite() http.SameSite {
	switch strings.ToLower(os.Getenv("COOKIE_SAMESITE")) {
	case "none":
		return http.SameSiteNoneMode
	case "strict":
		return http.SameSiteStrictMode
	default:
		return http.SameSiteLaxMode
	}
}

type loginInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var in services.RegisterInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz form verisi", nil)
		return
	}

	_, err := h.svc.Register(in)
	if err != nil {
		if errors.Is(err, services.ErrEmailExists) {
			utils.Err(c, http.StatusConflict, "EMAIL_EXISTS", "Bu e-posta adresi zaten kullanımda", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	// Auto login the newly registered user
	loginRes, err := h.svc.Login(in.Email, in.Password)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", "Kayıt sonrası otomatik giriş başarısız", nil)
		return
	}

	h.setAuthCookies(c, loginRes.AccessToken, loginRes.RefreshToken)
	utils.Created(c, loginRes)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var in loginInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "E-posta ve şifre zorunludur", nil)
		return
	}

	res, err := h.svc.Login(in.Email, in.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCreds) {
			utils.Err(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Hatalı e-posta veya şifre", nil)
			return
		}
		if errors.Is(err, services.ErrUserInactive) {
			utils.Err(c, http.StatusForbidden, "USER_INACTIVE", "Kullanıcı hesabı aktif değil", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	h.setAuthCookies(c, res.AccessToken, res.RefreshToken)
	utils.OK(c, res)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	rawRt, err := c.Cookie("rt")
	if err != nil || rawRt == "" {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Refresh token bulunamadı", nil)
		return
	}

	res, err := h.svc.Refresh(rawRt)
	if err != nil {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum yenilenemedi", nil)
		return
	}

	h.setAuthCookies(c, res.AccessToken, res.RefreshToken)
	utils.OK(c, gin.H{"ok": true})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	rawRt, err := c.Cookie("rt")
	if err == nil && rawRt != "" {
		_ = h.svc.Logout(rawRt)
	}

	h.clearAuthCookies(c)
	utils.OK(c, gin.H{"ok": true})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Geçersiz kullanıcı", nil)
		return
	}

	res, err := h.svc.Me(userID)
	if err != nil {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı bulunamadı", nil)
		return
	}

	utils.OK(c, res)
}

func (h *AuthHandler) setAuthCookies(c *gin.Context, at, rt string) {
	c.SetSameSite(h.sameSite)
	// Access token cookie (~15m)
	c.SetCookie("at", at, 15*60, "/", "", h.cookieSecure, true)
	// Refresh token cookie (~7 days)
	c.SetCookie("rt", rt, 7*24*60*60, "/", "", h.cookieSecure, true)
}

func (h *AuthHandler) clearAuthCookies(c *gin.Context) {
	c.SetSameSite(h.sameSite)
	c.SetCookie("at", "", -1, "/", "", h.cookieSecure, true)
	c.SetCookie("rt", "", -1, "/", "", h.cookieSecure, true)
}

type changePasswordInput struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var in changePasswordInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Eski ve yeni şifre zorunludur. Yeni şifre en az 8 karakter olmalıdır.", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Geçersiz kullanıcı", nil)
		return
	}

	err = h.svc.ChangePassword(userID, in.OldPassword, in.NewPassword)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCreds) {
			utils.Err(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Eski şifreniz hatalı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"message": "Şifreniz başarıyla değiştirildi"})
}
