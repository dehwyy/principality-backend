package ds

import (
	"github.com/golang-jwt/jwt"

	"github.com/dehwyy/principality-backend/internal/app/role"
)

type JWTClaims struct {
	jwt.StandardClaims
	ArchaeologistID uint      `json:"archaeologist_id"`
	Role            role.Role `json:"role"`
}
