package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/dehwyy/principality-backend/internal/app/ds"
)

func (r *Repository) GetArchaeologistByLogin(archaeologistLogin string) (*ds.Archaeologist, error) {
	var archaeologist ds.Archaeologist
	err := r.db.
		Where("archaeologist_login = ?", archaeologistLogin).
		First(&archaeologist).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &archaeologist, nil
}

func (r *Repository) AddArchaeologist(archaeologist *ds.Archaeologist) error {
	err := r.db.Create(archaeologist).Error
	if err != nil {
		return fmt.Errorf("ошибка при добавлении археолога: %w", err)
	}
	return nil
}
