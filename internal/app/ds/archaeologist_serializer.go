package ds

type ArchaeologistRegistration struct {
	ArchaeologistLogin    string `json:"archaeologist_login" binding:"required,max=25"`
	ArchaeologistPassword string `json:"archaeologist_password" binding:"required,max=100"`
}

type ArchaeologistCredentials struct {
	ArchaeologistLogin    string `json:"archaeologist_login" binding:"required"`
	ArchaeologistPassword string `json:"archaeologist_password" binding:"required"`
}
