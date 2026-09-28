package ds

import "time"

const (
	PrincipalityStatusDraft     = "draft"
	PrincipalityStatusPublished = "published"
	PrincipalityStatusRemoved   = "removed"
)

type Principality struct {
	PrincipalityID      uint       `gorm:"primaryKey" json:"principality_id"`
	PrincipalityName    string     `gorm:"type:varchar(120);not null" json:"principality_name"`
	PrincipalitySummary string     `gorm:"type:varchar(600)" json:"principality_summary"`
	PrincipalityStatus  string     `gorm:"type:varchar(16);not null;default:'draft'" json:"principality_status"`
	ImageKey            string     `gorm:"type:varchar(80);not null" json:"image_key"`
	VideoKey            string     `gorm:"type:varchar(80);not null" json:"video_key"`
	FoundingDate        *time.Time `gorm:"type:date" json:"founding_date"`
	LandCoefficient     *float64   `gorm:"type:numeric(4,2)" json:"land_coefficient"`
	CreatedAt           time.Time  `gorm:"not null" json:"created_at"`
	PublishedAt         *time.Time `gorm:"default:null" json:"published_at"`
	CreatedBy           uint       `gorm:"not null" json:"created_by"`

	Creator Archaeologist `gorm:"foreignKey:CreatedBy;constraint:OnUpdate:NO ACTION,OnDelete:NO ACTION" json:"-"`
}

func (Principality) TableName() string {
	return "principality"
}
