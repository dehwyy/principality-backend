package repository

import (
	"errors"
	"sort"
)

var ErrPrincipalityNotFound = errors.New("княжество не найдено")

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

type Principality struct {
	PrincipalityID         int
	PrincipalityName       string
	PrincipalitySummary    string
	PrincipalityStatus     PrincipalityStatus
	SettlementAreaHectares float64
	SettlementType         SettlementType
	ImageKey               string
	VideoKey               string
	AreaSource             string
	LikedBy                []int
}

func (n Principality) LikeCount() int {
	return len(n.LikedBy)
}

func (n Principality) IsEstimated() bool {
	return n.AreaSource == ""
}

type Repository struct {
	principalities []Principality
}

func NewRepository() (*Repository, error) {
	return &Repository{
		principalities: principalities(),
	}, nil
}

func (r *Repository) GetPublishedPrincipalities(minArea float64) []Principality {
	published := make([]Principality, 0, len(r.principalities))
	for _, principality := range r.principalities {
		if principality.PrincipalityStatus != PrincipalityStatusPublished {
			continue
		}
		if minArea > 0 && principality.SettlementAreaHectares < minArea {
			continue
		}
		published = append(published, principality)
	}
	sort.Slice(published, func(i, j int) bool {
		return published[i].PrincipalityID < published[j].PrincipalityID
	})
	return published
}

func (r *Repository) GetPublishedPrincipality(principalityID int) (Principality, error) {
	for _, principality := range r.principalities {
		if principality.PrincipalityID == principalityID && principality.PrincipalityStatus == PrincipalityStatusPublished {
			return principality, nil
		}
	}
	return Principality{}, ErrPrincipalityNotFound
}

func (r *Repository) GetNextPublishedPrincipality(afterID int) (Principality, error) {
	published := r.GetPublishedPrincipalities(0)
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

func (r *Repository) GetFirstPublishedPrincipalityID() int {
	published := r.GetPublishedPrincipalities(0)
	if len(published) == 0 {
		return 0
	}
	return published[0].PrincipalityID
}

func (r *Repository) GetDraftPrincipality() (Principality, error) {
	for _, principality := range r.principalities {
		if principality.PrincipalityStatus == PrincipalityStatusDraft {
			return principality, nil
		}
	}
	return Principality{}, ErrPrincipalityNotFound
}

func principalities() []Principality {
	const kuza = "https://archaeolog.ru/el-bib/el-cat/el-books/el-books-1996/kuza-1996"
	const tikhomirov = "http://rusarch.ru/tihomirov1.htm"

	return []Principality{
		{
			PrincipalityID:         1,
			PrincipalityName:       "Киевское княжество",
			PrincipalitySummary:    "Стольный град Киев - крупнейший укреплённый центр Древней Руси. Город Владимира, город Ярослава и город Изяслава вместе с Подолом дают около 300 гектаров укреплённой площади, что делает Киев верхней границей шкалы при любой оценке городского населения домонгольской Руси.",
			PrincipalityStatus:     PrincipalityStatusPublished,
			SettlementAreaHectares: 300,
			SettlementType:         SettlementCapital,
			ImageKey:               "kiev.jpg",
			VideoKey:               "kiev.mp4",
			AreaSource:             tikhomirov,
			LikedBy:                []int{1, 3, 5, 7},
		},
		{
			PrincipalityID:         2,
			PrincipalityName:       "Черниговское княжество",
			PrincipalitySummary:    "Чернигов с детинцем, окольным градом и предградьем занимал около 160 гектаров. Второй по величине центр Южной Руси: княжеский двор, Спасский собор и обширный посад вдоль Десны формировали плотную усадебную застройку.",
			PrincipalityStatus:     PrincipalityStatusPublished,
			SettlementAreaHectares: 160,
			SettlementType:         SettlementCapital,
			ImageKey:               "chernigov.jpg",
			VideoKey:               "chernigov.mp4",
			AreaSource:             tikhomirov,
			LikedBy:                []int{2, 4, 6},
		},
		{
			PrincipalityID:         3,
			PrincipalityName:       "Псковская земля",
			PrincipalitySummary:    "Псков с Кромом, Довмонтовым городом и посадом - около 150 гектаров укреплённой площади. Каменные стены строились в несколько очередей, поэтому площадь заметно менялась по столетиям, и для расчёта берётся состояние на конец домонгольского периода.",
			PrincipalityStatus:     PrincipalityStatusPublished,
			SettlementAreaHectares: 150,
			SettlementType:         SettlementCapital,
			ImageKey:               "pskov.jpg",
			VideoKey:               "pskov.mp4",
			AreaSource:             tikhomirov,
			LikedBy:                []int{1, 8},
		},
		{
			PrincipalityID:         4,
			PrincipalityName:       "Владимиро-Суздальское княжество",
			PrincipalitySummary:    "Владимир-на-Клязьме при Андрее Боголюбском вырос до 145 гектаров: Печерний город, Новый город и Ветчаной город с валами и Золотыми воротами. Столица Северо-Восточной Руси, застройка усадебная и плотная.",
			PrincipalityStatus:     PrincipalityStatusPublished,
			SettlementAreaHectares: 145,
			SettlementType:         SettlementCapital,
			ImageKey:               "vladimir.jpg",
			VideoKey:               "vladimir.mp4",
			AreaSource:             tikhomirov,
			LikedBy:                []int{3, 5},
		},
		{
			PrincipalityID:         5,
			PrincipalityName:       "Смоленское княжество",
			PrincipalitySummary:    "Смоленск занимал около 100 гектаров на днепровских холмах. Ключевой пункт пути из варяг в греки, с развитым ремесленным посадом и десятками каменных храмов домонгольской постройки.",
			PrincipalityStatus:     PrincipalityStatusPublished,
			SettlementAreaHectares: 100,
			SettlementType:         SettlementCapital,
			ImageKey:               "smolensk.jpg",
			VideoKey:               "smolensk.mp4",
			AreaSource:             tikhomirov,
			LikedBy:                []int{2, 7, 9},
		},
		{
			PrincipalityID:         6,
			PrincipalityName:       "Переславль-Залесское княжество",
			PrincipalitySummary:    "Переславль-Залесский основан Юрием Долгоруким в 1152 году: земляной вал длиной около двух с половиной километров охватывает примерно 40 гектаров. Валы сохранились по всему периметру, поэтому площадь измеряется не по реконструкции, а по существующему укреплению.",
			PrincipalityStatus:     PrincipalityStatusPublished,
			SettlementAreaHectares: 40,
			SettlementType:         SettlementHillfort,
			ImageKey:               "pereslavl_zalessky.jpg",
			VideoKey:               "pereslavl_zalessky.mp4",
			AreaSource:             tikhomirov,
			LikedBy:                []int{4},
		},
		{
			PrincipalityID:         7,
			PrincipalityName:       "Переяславское княжество",
			PrincipalitySummary:    "Переяславль Южный на Трубеже - около 80 гектаров. Третий стол Русской земли наряду с Киевом и Черниговом, но постоянные половецкие набеги держали застройку менее плотной, чем в глубине страны.",
			PrincipalityStatus:     PrincipalityStatusPublished,
			SettlementAreaHectares: 80,
			SettlementType:         SettlementFortifiedTown,
			ImageKey:               "pereyaslavl.jpg",
			VideoKey:               "pereyaslavl.mp4",
			AreaSource:             tikhomirov,
			LikedBy:                []int{6, 8},
		},
		{
			PrincipalityID:         8,
			PrincipalityName:       "Полоцкое княжество",
			PrincipalitySummary:    "Полоцк на Западной Двине - около 58 гектаров. Древнейший центр кривичей, рано обособившийся от Киева; Софийский собор и детинец на высоком берегу задают компактную планировку.",
			PrincipalityStatus:     PrincipalityStatusPublished,
			SettlementAreaHectares: 58,
			SettlementType:         SettlementFortifiedTown,
			ImageKey:               "polotsk.jpg",
			VideoKey:               "polotsk.mp4",
			AreaSource:             kuza,
			LikedBy:                []int{9},
		},
		{
			PrincipalityID:         9,
			PrincipalityName:       "Галицкое княжество",
			PrincipalitySummary:    "",
			PrincipalityStatus:     PrincipalityStatusDraft,
			SettlementAreaHectares: 0,
			SettlementType:         "",
			ImageKey:               "galich.jpg",
			VideoKey:               "galich.mp4",
			AreaSource:             kuza,
			LikedBy:                []int{},
		},
		{
			PrincipalityID:         10,
			PrincipalityName:       "Муромо-Рязанское княжество",
			PrincipalitySummary:    "Старая Рязань на Оке - около 53 гектаров. Столица княжества, разорённая в 1237 году и после этого не восстановленная, поэтому её планировка изучена археологически лучше многих действующих городов.",
			PrincipalityStatus:     PrincipalityStatusRemoved,
			SettlementAreaHectares: 53,
			SettlementType:         SettlementFortifiedTown,
			ImageKey:               "ryazan.jpg",
			VideoKey:               "ryazan.mp4",
			AreaSource:             kuza,
			LikedBy:                []int{5},
		},
	}
}
