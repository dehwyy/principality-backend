package repository

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
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

type SettlementType string

const (
	SettlementCapital       SettlementType = "capital"
	SettlementFortifiedTown SettlementType = "fortified_town"
	SettlementHillfort      SettlementType = "hillfort"
)

var settlementTitles = map[SettlementType]string{
	SettlementCapital:       "стольный град",
	SettlementFortifiedTown: "окольный город",
	SettlementHillfort:      "городище-крепость",
}

var builtUpRatios = map[SettlementType]float64{
	SettlementCapital:       0.55,
	SettlementFortifiedTown: 0.45,
	SettlementHillfort:      0.30,
}

func (t SettlementType) String() string {
	return string(t)
}

func (t SettlementType) Title() string {
	if title, ok := settlementTitles[t]; ok {
		return title
	}
	return string(t)
}

func (t SettlementType) BuiltUpRatio() float64 {
	if ratio, ok := builtUpRatios[t]; ok {
		return ratio
	}
	return builtUpRatios[SettlementHillfort]
}

type SettlementAreaHectares float64

func SettlementAreaHectaresFromString(
	s string,
) SettlementAreaHectares {
	ha, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return SettlementAreaHectares(ha)
}

func (s SettlementAreaHectares) String() string {
	if s == 0 {
		return ""
	}
	return strings.Replace(
		strings.TrimSuffix(strings.TrimRight(fmt.Sprintf("%.1f", s), "0"), "."),
		".",
		",",
		1,
	)
}

type Principality struct {
	PrincipalityID         PrincipalityID
	PrincipalityName       string
	PrincipalitySummary    string
	PrincipalityStatus     PrincipalityStatus
	SettlementAreaHectares SettlementAreaHectares
	SettlementType         SettlementType
	ImageKey               string
	VideoKey               string
	LikedBy                []int
}

func (n Principality) LikeCount() int {
	return len(n.LikedBy)
}
