package ds

type PrincipalitySerializer struct {
	Principality
	PrincipalityImageURL string `json:"principality_image_url"`
	PrincipalityVideoURL string `json:"principality_video_url"`
}

type PrincipalityCatalogSerializer struct {
	PrincipalitySerializer
	CreatedByCurrentArchaeologist int   `json:"created_by_current_archaeologist"`
	PrincipalityLikeCount         int64 `json:"principality_like_count"`
}

type PrincipalityFeedSerializer struct {
	PrincipalitySerializer
	PrincipalityLikeCount int64 `json:"principality_like_count"`
	NextPrincipalityID    uint  `json:"next_principality_id"`
}

type PrincipalityLikeSerializer struct {
	PrincipalityID        uint  `json:"principality_id"`
	PrincipalityLiked     int   `json:"principality_liked"`
	PrincipalityLikeCount int64 `json:"principality_like_count"`
}

type PrincipalityPublication struct {
	PrincipalitySummary string  `json:"principality_summary" binding:"required"`
	FoundingDate        string  `json:"founding_date" binding:"required"`
	LandCoefficient     float64 `json:"land_coefficient" binding:"required"`
}

type PrincipalityLikeMark struct {
	PrincipalityLiked *int `json:"principality_liked" binding:"required"`
}
