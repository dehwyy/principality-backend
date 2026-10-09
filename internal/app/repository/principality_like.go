package repository

import (
	"github.com/dehwyy/principality-backend/internal/app/ds"
)

func (r *Repository) SetPrincipalityLike(principalityID uint, archaeologistID uint) error {
	var principalityLike ds.PrincipalityLike
	err := r.db.
		Where(ds.PrincipalityLike{ArchaeologistRef: archaeologistID, PrincipalityRef: principalityID}).
		FirstOrCreate(&principalityLike).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) UnsetPrincipalityLike(principalityID uint, archaeologistID uint) error {
	err := r.db.
		Where("archaeologist_id = ? AND principality_id = ?", archaeologistID, principalityID).
		Delete(&ds.PrincipalityLike{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) IsPrincipalityLikedBy(principalityID uint, archaeologistID uint) (bool, error) {
	var count int64
	err := r.db.Model(&ds.PrincipalityLike{}).
		Where("archaeologist_id = ? AND principality_id = ?", archaeologistID, principalityID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
