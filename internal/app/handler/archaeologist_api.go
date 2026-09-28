package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dehwyy/principality-backend/internal/app/ds"
)

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
		ArchaeologistPassword: registration.ArchaeologistPassword,
	}
	err = h.Repository.AddArchaeologist(&archaeologist)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, archaeologist)
}

func (h *Handler) LoginArchaeologistAPI(ctx *gin.Context) {
	var credentials ds.ArchaeologistCredentials
	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "аутентификация будет реализована в лабораторной 4",
	})
}

func (h *Handler) LogoutArchaeologistAPI(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "деавторизация будет реализована в лабораторной 4",
	})
}
