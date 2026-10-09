package handler

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	log "github.com/sirupsen/logrus"

	"github.com/dehwyy/principality-backend/internal/app/ds"
)

// RegisterArchaeologistAPI godoc
// @Summary      Регистрация археолога
// @Description  Создаёт археолога с логином и паролем, пароль хранится как SHA-1
// @Tags         Археологи
// @Accept       json
// @Produce      json
// @Param        registration  body      ds.ArchaeologistRegistration  true  "Логин и пароль археолога"
// @Success      201           {object}  ds.Archaeologist
// @Failure      400           {object}  map[string]string
// @Failure      409           {object}  map[string]string
// @Router       /api/archaeologists [post]
func (h *Handler) RegisterArchaeologistAPI(ctx *gin.Context) {
	var registration ds.ArchaeologistRegistration
	if err := ctx.ShouldBindJSON(&registration); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	existingArchaeologist, err := h.Repository.GetArchaeologistByLogin(registration.ArchaeologistLogin)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if existingArchaeologist != nil {
		h.errorHandler(ctx, http.StatusConflict, errors.New("логин археолога уже занят"))
		return
	}

	archaeologist := ds.Archaeologist{
		ArchaeologistLogin:    registration.ArchaeologistLogin,
		ArchaeologistPassword: generateHashString(registration.ArchaeologistPassword),
	}
	err = h.Repository.AddArchaeologist(&archaeologist)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, archaeologist)
}

// LoginArchaeologistAPI godoc
// @Summary      Аутентификация археолога
// @Description  Проверяет логин и пароль и выдаёт JWT для заголовка Authorization
// @Tags         Археологи
// @Accept       json
// @Produce      json
// @Param        credentials  body      ds.ArchaeologistCredentials  true  "Логин и пароль археолога"
// @Success      200          {object}  ds.ArchaeologistToken
// @Failure      400
// @Failure      403
// @Router       /api/archaeologists/login [post]
func (h *Handler) LoginArchaeologistAPI(gCtx *gin.Context) {
	cfg := h.Config
	req := &ds.ArchaeologistCredentials{}

	err := json.NewDecoder(gCtx.Request.Body).Decode(req)
	if err != nil {
		gCtx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	archaeologist, err := h.Repository.GetArchaeologistByLogin(req.ArchaeologistLogin)
	if err != nil {
		gCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if archaeologist != nil && req.ArchaeologistLogin == archaeologist.ArchaeologistLogin &&
		archaeologist.ArchaeologistPassword == generateHashString(req.ArchaeologistPassword) {
		token := jwt.NewWithClaims(cfg.JWT.SigningMethod, &ds.JWTClaims{
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(cfg.JWT.ExpiresIn).Unix(),
				IssuedAt:  time.Now().Unix(),
				Issuer:    "principality-web",
			},
			ArchaeologistID: archaeologist.ArchaeologistID,
			Role:            archaeologist.ArchaeologistRole(),
		})
		if token == nil {
			gCtx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("token is nil"))
			return
		}

		strToken, err := token.SignedString([]byte(cfg.JWT.Token))
		if err != nil {
			gCtx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("cant create str token"))
			return
		}

		gCtx.JSON(http.StatusOK, ds.ArchaeologistToken{
			ExpiresIn:   cfg.JWT.ExpiresIn,
			AccessToken: strToken,
			TokenType:   "Bearer",
		})
		return
	}

	gCtx.AbortWithStatus(http.StatusForbidden)
}

// LogoutArchaeologistAPI godoc
// @Summary      Деавторизация археолога
// @Description  Кладёт JWT из заголовка Authorization в blacklist Redis
// @Tags         Археологи
// @Security     BearerAuth
// @Success      200
// @Failure      400
// @Router       /api/archaeologists/logout [post]
func (h *Handler) LogoutArchaeologistAPI(gCtx *gin.Context) {
	jwtStr := gCtx.GetHeader("Authorization")
	if !strings.HasPrefix(jwtStr, jwtPrefix) {
		gCtx.AbortWithStatus(http.StatusBadRequest)

		return
	}

	jwtStr = jwtStr[len(jwtPrefix):]

	_, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.Config.JWT.Token), nil
	})
	if err != nil {
		gCtx.AbortWithError(http.StatusBadRequest, err)
		log.Println(err)

		return
	}

	err = h.Redis.WriteJWTToBlacklist(gCtx.Request.Context(), jwtStr, h.Config.JWT.ExpiresIn)
	if err != nil {
		gCtx.AbortWithError(http.StatusInternalServerError, err)

		return
	}

	gCtx.Status(http.StatusOK)
}

func generateHashString(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
