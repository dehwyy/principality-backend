package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

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

func (h *Handler) AddPrincipalityAPI(ctx *gin.Context) {
	err := ctx.Request.ParseMultipartForm(32 << 20)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	principalityName := strings.TrimSpace(ctx.Request.FormValue("principality_name"))
	if principalityName == "" {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("название княжества не заполнено"))
		return
	}

	currentArchaeologistID := auth.CurrentArchaeologist().ArchaeologistID
	existingDraft, err := h.Repository.GetDraftPrincipality(currentArchaeologistID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if existingDraft != nil {
		h.errorHandler(ctx, http.StatusConflict, errors.New("у археолога уже есть черновик княжества"))
		return
	}

	imageHeader, err := ctx.FormFile("principality_image")
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("нужен файл изображения княжества"))
		return
	}
	imageType, code, err := validateFileUpload(imageHeader, isImage)
	if err != nil {
		h.errorHandler(ctx, code, err)
		return
	}

	videoHeader, err := ctx.FormFile("principality_video")
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("нужен файл видео княжества"))
		return
	}
	videoType, code, err := validateFileUpload(videoHeader, isVideo)
	if err != nil {
		h.errorHandler(ctx, code, err)
		return
	}

	imageKey, err := repository.NewPrincipalityMediaKey(imageType)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	videoKey, err := repository.NewPrincipalityMediaKey(videoType)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	err = h.Repository.UploadPrincipalityMedia(imageHeader, imageKey)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	err = h.Repository.UploadPrincipalityMedia(videoHeader, videoKey)
	if err != nil {
		h.removePrincipalityMedia(imageKey)
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	principality := ds.Principality{
		PrincipalityName:   principalityName,
		PrincipalityStatus: ds.PrincipalityStatusDraft,
		ImageKey:           imageKey,
		VideoKey:           videoKey,
		CreatedAt:          time.Now(),
		CreatedBy:          currentArchaeologistID,
	}
	err = h.Repository.AddPrincipality(&principality)
	if err != nil {
		h.removePrincipalityMedia(imageKey, videoKey)
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, h.principalitySerializer(&principality))
}

func (h *Handler) PublishPrincipalityAPI(ctx *gin.Context) {
	principalityID, ok := h.parsePrincipalityID(ctx)
	if !ok {
		return
	}

	principality, err := h.Repository.GetPrincipality(principalityID)
	if err != nil {
		if errors.Is(err, repository.ErrPrincipalityNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
			return
		}
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	currentArchaeologistID := auth.CurrentArchaeologist().ArchaeologistID
	if principality.CreatedBy != currentArchaeologistID {
		h.errorHandler(ctx, http.StatusForbidden, errors.New("княжество создано другим археологом"))
		return
	}
	if principality.PrincipalityStatus != ds.PrincipalityStatusDraft {
		h.errorHandler(ctx, http.StatusConflict, errors.New("опубликовать можно только черновик княжества"))
		return
	}

	var publication ds.PrincipalityPublication
	if err := ctx.ShouldBindJSON(&publication); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	principalitySummary := strings.TrimSpace(publication.PrincipalitySummary)
	if principalitySummary == "" {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("описание княжества не заполнено"))
		return
	}
	foundingDate, err := time.Parse(foundingDateLayout, publication.FoundingDate)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("дата основания княжества должна быть в формате ГГГГ-ММ-ДД"))
		return
	}
	if publication.LandCoefficient <= 0 || publication.LandCoefficient > 1 {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("коэффициент земли должен быть больше 0 и не больше 1"))
		return
	}

	err = h.Repository.PublishPrincipality(
		principalityID,
		currentArchaeologistID,
		principalitySummary,
		foundingDate,
		publication.LandCoefficient,
	)
	if err != nil {
		if errors.Is(err, repository.ErrPrincipalityNotFound) {
			h.errorHandler(ctx, http.StatusConflict, errors.New("опубликовать можно только черновик княжества"))
			return
		}
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	publishedPrincipality, err := h.Repository.GetPrincipality(principalityID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, h.principalitySerializer(publishedPrincipality))
}

func (h *Handler) DeletePrincipalityAPI(ctx *gin.Context) {
	principalityID, ok := h.parsePrincipalityID(ctx)
	if !ok {
		return
	}

	principality, err := h.Repository.GetPrincipality(principalityID)
	if err != nil {
		if errors.Is(err, repository.ErrPrincipalityNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
			return
		}
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	currentArchaeologistID := auth.CurrentArchaeologist().ArchaeologistID
	if principality.CreatedBy != currentArchaeologistID {
		h.errorHandler(ctx, http.StatusForbidden, errors.New("княжество создано другим археологом"))
		return
	}

	err = h.Repository.RemovePrincipalityByArchaeologist(principalityID, currentArchaeologistID)
	if err != nil {
		if errors.Is(err, repository.ErrPrincipalityNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
			return
		}
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "княжество удалено",
	})
}

func (h *Handler) LikePrincipalityAPI(ctx *gin.Context) {
	principalityID, ok := h.parsePrincipalityID(ctx)
	if !ok {
		return
	}

	var likeMark ds.PrincipalityLikeMark
	if err := ctx.ShouldBindJSON(&likeMark); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	principalityLiked := *likeMark.PrincipalityLiked
	if principalityLiked != 0 && principalityLiked != 1 {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("principality_liked принимает 0 или 1"))
		return
	}

	principality, err := h.Repository.GetPrincipality(principalityID)
	if err != nil {
		if errors.Is(err, repository.ErrPrincipalityNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
			return
		}
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if principality.PrincipalityStatus != ds.PrincipalityStatusPublished {
		h.errorHandler(ctx, http.StatusNotFound, repository.ErrPrincipalityNotFound)
		return
	}

	currentArchaeologistID := auth.CurrentArchaeologist().ArchaeologistID
	if principalityLiked == 1 {
		err = h.Repository.SetPrincipalityLike(principalityID, currentArchaeologistID)
	} else {
		err = h.Repository.UnsetPrincipalityLike(principalityID, currentArchaeologistID)
	}
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	likeCount, err := h.Repository.CountPrincipalityLikesByID(principalityID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, ds.PrincipalityLikeSerializer{
		PrincipalityID:        principalityID,
		PrincipalityLiked:     principalityLiked,
		PrincipalityLikeCount: likeCount,
	})
}

func (h *Handler) removePrincipalityMedia(mediaKeys ...string) {
	for _, mediaKey := range mediaKeys {
		err := h.Repository.RemovePrincipalityMedia(mediaKey)
		if err != nil {
			logrus.Error(err.Error())
		}
	}
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
