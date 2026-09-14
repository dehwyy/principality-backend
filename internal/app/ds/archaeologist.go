package ds

type Archaeologist struct {
	ArchaeologistID       uint   `gorm:"primaryKey"`
	ArchaeologistLogin    string `gorm:"type:varchar(25);unique;not null"`
	ArchaeologistPassword string `gorm:"type:varchar(100);not null"`
	IsChronicleKeeper     bool   `gorm:"type:boolean;default:false"`
}

func (Archaeologist) TableName() string {
	return "archaeologist"
}
