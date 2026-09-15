package ds

import "time"

const (
	PrincipalityStatusDraft     = "draft"
	PrincipalityStatusPublished = "published"
	PrincipalityStatusRemoved   = "removed"
)

type Principality struct {
	PrincipalityID      uint       `gorm:"primaryKey"`
	PrincipalityName    string     `gorm:"type:varchar(120);not null"`
	PrincipalitySummary string     `gorm:"type:varchar(600)"`
	PrincipalityStatus  string     `gorm:"type:varchar(16);not null;default:'draft'"`
	ImageKey            string     `gorm:"type:varchar(80)"`
	VideoKey            string     `gorm:"type:varchar(80)"`
	FoundingDate        *time.Time `gorm:"type:date"`
	LandCoefficient     *float64   `gorm:"type:numeric(4,2)"`
	CreatedAt           time.Time  `gorm:"not null"`
	PublishedAt         *time.Time `gorm:"default:null"`
	CreatedBy           uint       `gorm:"not null"`

	Creator Archaeologist `gorm:"foreignKey:CreatedBy;constraint:OnUpdate:NO ACTION,OnDelete:NO ACTION"`
}

func (Principality) TableName() string {
	return "principality"
}
