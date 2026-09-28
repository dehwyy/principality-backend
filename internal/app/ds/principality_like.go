package ds

type PrincipalityLike struct {
	PrincipalityLikeID uint `gorm:"primaryKey" json:"principality_like_id"`
	ArchaeologistRef   uint `gorm:"column:archaeologist_id;not null;uniqueIndex:idx_principality_like" json:"archaeologist_id"`
	PrincipalityRef    uint `gorm:"column:principality_id;not null;uniqueIndex:idx_principality_like" json:"principality_id"`

	Archaeologist Archaeologist `gorm:"foreignKey:ArchaeologistRef;constraint:OnUpdate:NO ACTION,OnDelete:NO ACTION" json:"-"`
	Principality  Principality  `gorm:"foreignKey:PrincipalityRef;constraint:OnUpdate:NO ACTION,OnDelete:NO ACTION" json:"-"`
}

func (PrincipalityLike) TableName() string {
	return "principality_like"
}
