package ds

type PrincipalityLike struct {
	PrincipalityLikeID uint `gorm:"primaryKey"`
	ArchaeologistRef   uint `gorm:"column:archaeologist_id;not null;uniqueIndex:idx_principality_like"`
	PrincipalityRef    uint `gorm:"column:principality_id;not null;uniqueIndex:idx_principality_like"`

	Archaeologist Archaeologist `gorm:"foreignKey:ArchaeologistRef;constraint:OnUpdate:NO ACTION,OnDelete:NO ACTION"`
	Principality  Principality  `gorm:"foreignKey:PrincipalityRef;constraint:OnUpdate:NO ACTION,OnDelete:NO ACTION"`
}

func (PrincipalityLike) TableName() string {
	return "principality_like"
}
