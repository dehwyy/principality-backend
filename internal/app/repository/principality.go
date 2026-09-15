package repository

import (
	"database/sql"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/dehwyy/principality-backend/internal/app/ds"
)

var ErrPrincipalityNotFound = errors.New("княжество не найдено")

func (r *Repository) GetPublishedPrincipalities(foundedBefore *time.Time) ([]ds.Principality, error) {
	var principalities []ds.Principality
	publishedPrincipalities := r.db.Where("principality_status = ?", ds.PrincipalityStatusPublished)
	if foundedBefore != nil {
		publishedPrincipalities = publishedPrincipalities.Where("founding_date <= ?", *foundedBefore)
	}
	err := publishedPrincipalities.Order("principality_id").Find(&principalities).Error
	if err != nil {
		return nil, err
	}
	return principalities, nil
}

func (r *Repository) GetPublishedPrincipality(principalityID uint) (*ds.Principality, error) {
	principalitySelect := `SELECT principality_id, principality_name, principality_summary, principality_status,
		image_key, video_key, founding_date, land_coefficient, created_at, published_at, created_by
		FROM principality WHERE principality_id = $1 AND principality_status = $2`

	row := r.db.Raw(principalitySelect, principalityID, ds.PrincipalityStatusPublished).Row()

	principality := &ds.Principality{}

	err := row.Scan(
		&principality.PrincipalityID,
		&principality.PrincipalityName,
		&principality.PrincipalitySummary,
		&principality.PrincipalityStatus,
		&principality.ImageKey,
		&principality.VideoKey,
		&principality.FoundingDate,
		&principality.LandCoefficient,
		&principality.CreatedAt,
		&principality.PublishedAt,
		&principality.CreatedBy,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPrincipalityNotFound
		}
		return nil, err
	}

	return principality, nil
}

func (r *Repository) GetFirstPublishedPrincipality() (*ds.Principality, error) {
	var principality ds.Principality
	err := r.db.
		Where("principality_status = ?", ds.PrincipalityStatusPublished).
		Order("principality_id").
		First(&principality).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPrincipalityNotFound
		}
		return nil, err
	}
	return &principality, nil
}

func (r *Repository) GetNextPublishedPrincipality(afterID uint) (*ds.Principality, error) {
	var principality ds.Principality
	err := r.db.
		Where("principality_status = ? AND principality_id > ?", ds.PrincipalityStatusPublished, afterID).
		Order("principality_id").
		First(&principality).Error
	if err == nil {
		return &principality, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return r.GetFirstPublishedPrincipality()
}

func (r *Repository) GetDraftPrincipality(archaeologistID uint) (*ds.Principality, error) {
	var principality ds.Principality
	err := r.db.
		Where("principality_status = ? AND created_by = ?", ds.PrincipalityStatusDraft, archaeologistID).
		First(&principality).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &principality, nil
}

func (r *Repository) CreatePrincipalityDraft(principalityName string, archaeologistID uint) (*ds.Principality, error) {
	principality := &ds.Principality{
		PrincipalityName:   principalityName,
		PrincipalityStatus: ds.PrincipalityStatusDraft,
		CreatedAt:          time.Now(),
		CreatedBy:          archaeologistID,
	}
	err := r.db.Create(principality).Error
	if err != nil {
		return nil, err
	}
	return principality, nil
}

func (r *Repository) PublishPrincipality(
	principalityID uint,
	archaeologistID uint,
	principalitySummary string,
	foundingDate time.Time,
	landCoefficient float64,
) error {
	result := r.db.Model(&ds.Principality{}).
		Where("principality_id = ? AND created_by = ? AND principality_status = ?",
			principalityID, archaeologistID, ds.PrincipalityStatusDraft).
		Updates(map[string]any{
			"principality_summary": principalitySummary,
			"founding_date":        foundingDate,
			"land_coefficient":     landCoefficient,
			"principality_status":  ds.PrincipalityStatusPublished,
			"published_at":         time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrPrincipalityNotFound
	}
	return nil
}

func (r *Repository) RemovePrincipality(principalityID uint) error {
	removalUpdate := `UPDATE principality SET principality_status = 'removed'
		WHERE principality_id = $1 AND principality_status <> 'removed'
		RETURNING principality_id`

	row := r.db.Raw(removalUpdate, principalityID).Row()

	var removedPrincipalityID uint
	err := row.Scan(&removedPrincipalityID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPrincipalityNotFound
		}
		return err
	}

	return nil
}

type principalityLikeCount struct {
	PrincipalityID uint
	LikeCount      int64
}

func (r *Repository) CountPrincipalityLikes() (map[uint]int64, error) {
	var counts []principalityLikeCount
	err := r.db.Model(&ds.PrincipalityLike{}).
		Select("principality_id, COUNT(*) AS like_count").
		Group("principality_id").
		Scan(&counts).Error
	if err != nil {
		return nil, err
	}
	likeCounts := make(map[uint]int64, len(counts))
	for _, count := range counts {
		likeCounts[count.PrincipalityID] = count.LikeCount
	}
	return likeCounts, nil
}

func (r *Repository) CountPrincipalityLikesByID(principalityID uint) (int64, error) {
	var count int64
	err := r.db.Model(&ds.PrincipalityLike{}).
		Where("principality_id = ?", principalityID).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
