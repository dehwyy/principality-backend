package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt"
	log "github.com/sirupsen/logrus"

	"github.com/dehwyy/principality-backend/internal/app/ds"
	"github.com/dehwyy/principality-backend/internal/app/role"
)

const jwtPrefix = "Bearer "

const archaeologistIDKey = "archaeologist_id"

func (h *Handler) WithAuthCheck(assignedRoles ...role.Role) func(ctx *gin.Context) {
	return func(gCtx *gin.Context) {
		jwtStr := gCtx.GetHeader("Authorization")
		if !strings.HasPrefix(jwtStr, jwtPrefix) {
			gCtx.AbortWithStatus(http.StatusForbidden)

			return
		}

		jwtStr = jwtStr[len(jwtPrefix):]
		err := h.Redis.CheckJWTInBlacklist(gCtx.Request.Context(), jwtStr)
		if err == nil {
			gCtx.AbortWithStatus(http.StatusForbidden)

			return
		}
		if !errors.Is(err, redis.Nil) {
			gCtx.AbortWithError(http.StatusInternalServerError, err)

			return
		}

		token, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(h.Config.JWT.Token), nil
		})
		if err != nil {
			gCtx.AbortWithStatus(http.StatusForbidden)
			log.Println(err)

			return
		}

		myClaims := token.Claims.(*ds.JWTClaims)

		for _, oneOfAssignedRole := range assignedRoles {
			if myClaims.Role == oneOfAssignedRole {
				gCtx.Set(archaeologistIDKey, myClaims.ArchaeologistID)
				gCtx.Next()

				return
			}
		}
		gCtx.AbortWithStatus(http.StatusForbidden)
		log.Printf("role %d is not assigned in %v", myClaims.Role, assignedRoles)
	}
}

func (h *Handler) WithGuestAccess() func(ctx *gin.Context) {
	return func(gCtx *gin.Context) {
		jwtStr := gCtx.GetHeader("Authorization")
		if !strings.HasPrefix(jwtStr, jwtPrefix) {
			gCtx.Next()

			return
		}

		jwtStr = jwtStr[len(jwtPrefix):]
		err := h.Redis.CheckJWTInBlacklist(gCtx.Request.Context(), jwtStr)
		if !errors.Is(err, redis.Nil) {
			gCtx.Next()

			return
		}

		token, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(h.Config.JWT.Token), nil
		})
		if err == nil {
			gCtx.Set(archaeologistIDKey, token.Claims.(*ds.JWTClaims).ArchaeologistID)
		}
		gCtx.Next()
	}
}

func contextArchaeologistID(gCtx *gin.Context) uint {
	return gCtx.GetUint(archaeologistIDKey)
}
