package repository

type Repository struct {
	principalities []Principality
}

func NewRepository() (*Repository, error) {
	return &Repository{
		principalities: principalities(),
	}, nil
}

func (r *Repository) GetPublishedPrincipalities(
	foundedBefore FoundingDate,
) []Principality {
	published := make([]Principality, 0, len(r.principalities))
	for _, principality := range r.principalities {
		if principality.PrincipalityStatus != PrincipalityStatusPublished {
			continue
		}
		if !foundedBefore.IsZero() && principality.FoundingDate.After(foundedBefore) {
			continue
		}
		published = append(published, principality)
	}

	return published
}

func (r *Repository) GetPublishedPrincipality(
	principalityID PrincipalityID,
) (Principality, error) {
	for _, principality := range r.principalities {
		if principality.PrincipalityID == principalityID && principality.PrincipalityStatus == PrincipalityStatusPublished {
			return principality, nil
		}
	}
	return Principality{}, ErrPrincipalityNotFound
}

func (r *Repository) GetNextPublishedPrincipality(
	afterID PrincipalityID,
) (Principality, error) {
	published := r.GetPublishedPrincipalities(FoundingDate{})
	if len(published) == 0 {
		return Principality{}, ErrPrincipalityNotFound
	}
	for _, principality := range published {
		if principality.PrincipalityID > afterID {
			return principality, nil
		}
	}
	return published[0], nil
}

func (r *Repository) GetFirstPublishedPrincipality() (Principality, error) {
	published := r.GetPublishedPrincipalities(FoundingDate{})
	if len(published) == 0 {
		return Principality{}, ErrPrincipalityNotFound
	}
	return published[0], nil
}

func (r *Repository) GetDraftPrincipality() (Principality, error) {
	for _, principality := range r.principalities {
		if principality.PrincipalityStatus == PrincipalityStatusDraft {
			return principality, nil
		}
	}
	return Principality{}, ErrPrincipalityNotFound
}
