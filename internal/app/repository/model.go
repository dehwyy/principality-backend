package repository

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var ErrPrincipalityNotFound = errors.New("княжество не найдено")

type PrincipalityID int

func PrincipalityIDFromString(
	s string,
) (PrincipalityID, error) {
	id, err := strconv.Atoi(s)
	if err != nil {
		return PrincipalityID(0), err
	}
	return PrincipalityID(id), nil
}

func (id PrincipalityID) Int() int {
	return int(id)
}

type PrincipalityStatus string

const (
	PrincipalityStatusDraft     PrincipalityStatus = "draft"
	PrincipalityStatusPublished PrincipalityStatus = "published"
	PrincipalityStatusRemoved   PrincipalityStatus = "removed"
)

const foundingDateLayout = "2006-01-02"

type FoundingDate time.Time

func FoundingDateFromString(
	s string,
) (FoundingDate, error) {
	parsed, err := time.Parse(foundingDateLayout, s)
	if err != nil {
		return FoundingDate{}, err
	}
	return FoundingDate(parsed), nil
}

func FoundedInYear(year int) FoundingDate {
	return FoundingDate(time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC))
}

func (d FoundingDate) IsZero() bool {
	return time.Time(d).IsZero()
}

func (d FoundingDate) After(other FoundingDate) bool {
	return time.Time(d).After(time.Time(other))
}

func (d FoundingDate) String() string {
	if d.IsZero() {
		return ""
	}
	return time.Time(d).Format(foundingDateLayout)
}

func (d FoundingDate) YearTitle() string {
	if d.IsZero() {
		return ""
	}
	return fmt.Sprintf("%d г.", time.Time(d).Year())
}

type LandCoefficient float64

func (c LandCoefficient) String() string {
	if c == 0 {
		return ""
	}
	return strings.Replace(fmt.Sprintf("%.2f", float64(c)), ".", ",", 1)
}

func (c LandCoefficient) InputValue() string {
	if c == 0 {
		return ""
	}
	return fmt.Sprintf("%.2f", float64(c))
}

type Principality struct {
	PrincipalityID      PrincipalityID
	PrincipalityName    string
	PrincipalitySummary string
	PrincipalityStatus  PrincipalityStatus
	FoundingDate        FoundingDate
	LandCoefficient     LandCoefficient
	ImageKey            string
	VideoKey            string
	LikedBy             []int
}

func (n Principality) LikeCount() int {
	return len(n.LikedBy)
}
