package ds

import "time"

type ArchaeologistRegistration struct {
	ArchaeologistLogin    string `json:"archaeologist_login" binding:"required,max=25"`
	ArchaeologistPassword string `json:"archaeologist_password" binding:"required,max=100"`
}

type ArchaeologistCredentials struct {
	ArchaeologistLogin    string `json:"archaeologist_login" binding:"required"`
	ArchaeologistPassword string `json:"archaeologist_password" binding:"required"`
}

type ArchaeologistToken struct {
	ExpiresIn   time.Duration `json:"expires_in" swaggertype:"integer"`
	AccessToken string        `json:"access_token"`
	TokenType   string        `json:"token_type"`
}
