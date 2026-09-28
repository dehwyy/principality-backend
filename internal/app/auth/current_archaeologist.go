package auth

import (
	"sync"

	"github.com/dehwyy/principality-backend/internal/app/ds"
)

const (
	currentArchaeologistID    uint = 1
	currentArchaeologistLogin      = "a.kuza"
)

var (
	currentArchaeologistOnce sync.Once
	currentArchaeologist     *ds.Archaeologist
)

func CurrentArchaeologist() *ds.Archaeologist {
	currentArchaeologistOnce.Do(func() {
		currentArchaeologist = &ds.Archaeologist{
			ArchaeologistID:    currentArchaeologistID,
			ArchaeologistLogin: currentArchaeologistLogin,
		}
	})
	return currentArchaeologist
}
