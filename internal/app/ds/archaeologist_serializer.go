package ds

type ArchaeologistRegistration struct {
	ArchaeologistLogin    string `json:"archaeologist_login" binding:"required"`
	ArchaeologistPassword string `json:"archaeologist_password" binding:"required"`
}

type ArchaeologistCredentials struct {
	ArchaeologistLogin    string `json:"archaeologist_login" binding:"required"`
	ArchaeologistPassword string `json:"archaeologist_password" binding:"required"`
}
