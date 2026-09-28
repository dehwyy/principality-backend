package ds

type Archaeologist struct {
	ArchaeologistID       uint   `gorm:"primaryKey" json:"archaeologist_id"`
	ArchaeologistLogin    string `gorm:"type:varchar(25);unique;not null" json:"archaeologist_login"`
	ArchaeologistPassword string `gorm:"type:varchar(100);not null" json:"-"`
	IsChronicleKeeper     bool   `gorm:"type:boolean;default:false" json:"is_chronicle_keeper"`
}

func (Archaeologist) TableName() string {
	return "archaeologist"
}
