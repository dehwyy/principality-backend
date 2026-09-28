package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dehwyy/principality-backend/internal/app/auth"
	"github.com/dehwyy/principality-backend/internal/app/ds"
	"github.com/dehwyy/principality-backend/internal/app/repository"
)

func (h *Handler) GetPrincipalitiesAPI(ctx *gin.Context) {
	foundedBefore, _ := parseFoundedBefore(ctx)

	principalities, err := h.Repository.GetPublishedPrincipalities(foundedBefore)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	likeCounts, err := h.Repository.CountPrincipalityLikes()
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	currentArchaeologistID := auth.CurrentArchaeologist().ArchaeologistID
	catalog := make([]ds.PrincipalityCatalogSerializer, 0, len(principalities))
	for i := range principalities {
		createdByCurrentArchaeologist := 0
		if principalities[i].CreatedBy == currentArchaeologistID {
			createdByCurrentArchaeologist = 1
		}
		catalog = append(catalog, ds.PrincipalityCatalogSerializer{
			PrincipalitySerializer:        h.principalitySerializer(&principalities[i]),
			CreatedByCurrentArchaeologist: createdByCurrentArchaeologist,
			PrincipalityLikeCount:         likeCounts[principalities[i].PrincipalityID],
		})
	}

	ctx.JSON(http.StatusOK, catalog)
}

func (h *Handler) GetPrincipalityFeedAPI(ctx *gin.Context) {
	requestedPrincipality := ctx.Param("principalityId")

	var principality *ds.Principality
	var err error

	if requestedPrincipality == "" {
		principality, err = h.Repository.GetFirstPublishedPrincipality()
	} else {
		principalityID, ok := h.parsePrincipalityID(ctx)
		if !ok {
			return
		}
		if ctx.Query("next") == "true" {
			principality, err = h.Repository.GetNextPublishedPrincipality(principalityID)
		} else {
			principality, err = h.Repository.GetPublishedPrincipality(principalityID)
		}
	}
	if err != nil {
		if errors.Is(err, repository.ErrPrincipalityNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
			return
		}
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	likeCount, err := h.Repository.CountPrincipalityLikesByID(principality.PrincipalityID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	nextPrincipalityID := principality.PrincipalityID
	nextPrincipality, err := h.Repository.GetNextPublishedPrincipality(principality.PrincipalityID)
	if err == nil {
		nextPrincipalityID = nextPrincipality.PrincipalityID
	}

	ctx.JSON(http.StatusOK, ds.PrincipalityFeedSerializer{
		PrincipalitySerializer: h.principalitySerializer(principality),
		PrincipalityLikeCount:  likeCount,
		NextPrincipalityID:     nextPrincipalityID,
	})
}

func (h *Handler) GetPrincipalityDraftAPI(ctx *gin.Context) {
	principality, err := h.Repository.GetDraftPrincipality(auth.CurrentArchaeologist().ArchaeologistID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if principality == nil {
		h.errorHandler(ctx, http.StatusNotFound, errors.New("у археолога нет черновика княжества"))
		return
	}

	ctx.JSON(http.StatusOK, h.principalitySerializer(principality))
}

func (h *Handler) principalitySerializer(principality *ds.Principality) ds.PrincipalitySerializer {
	serializer := ds.PrincipalitySerializer{
		Principality: *principality,
	}
	if principality.ImageKey != "" {
		serializer.PrincipalityImageURL = h.principalityMediaURL(principality.ImageKey)
	}
	if principality.VideoKey != "" {
		serializer.PrincipalityVideoURL = h.principalityMediaURL(principality.VideoKey)
	}
	return serializer
}

func (h *Handler) principalityMediaURL(mediaKey string) string {
	return h.Config.MinioBaseURL + "/" + h.Config.MinioBucketName + "/" + mediaKey
}

func (h *Handler) parsePrincipalityID(ctx *gin.Context) (uint, bool) {
	principalityID, err := strconv.Atoi(ctx.Param("principalityId"))
	if err != nil || principalityID <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("идентификатор княжества должен быть положительным числом"))
		return 0, false
	}
	return uint(principalityID), true
}

func parseFoundedBefore(ctx *gin.Context) (*time.Time, string) {
	rawFoundedBefore := strings.TrimSpace(ctx.Query("foundedBefore"))
	foundedBefore, err := time.Parse(foundingDateLayout, rawFoundedBefore)
	if err != nil {
		return nil, ""
	}
	return &foundedBefore, rawFoundedBefore
}
