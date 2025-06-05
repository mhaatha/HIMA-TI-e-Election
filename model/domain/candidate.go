package domain

import "time"

type Candidate struct {
	Id                    int       `json:"id"`
	Number                int       `json:"number"`
	President             string    `json:"president"`
	Vice                  string    `json:"vice"`
	Vision                string    `json:"vision"`
	Mission               string    `json:"mission"`
	PhotoKey              string    `json:"photo_key"`
	PresidentStudyProgram string    `json:"president_study_program"`
	ViceStudyProgram      string    `json:"vice_study_program"`
	PresidentNIM          string    `json:"president_nim"`
	ViceNIM               string    `json:"vice_nim"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type CandidateWithURL struct {
	Id                    int       `json:"id"`
	Number                int       `json:"number"`
	President             string    `json:"president"`
	Vice                  string    `json:"vice"`
	Vision                string    `json:"vision"`
	Mission               string    `json:"mission"`
	PhotoURL              string    `json:"photo_url"`
	PresidentStudyProgram string    `json:"president_study_program"`
	ViceStudyProgram      string    `json:"vice_study_program"`
	PresidentNIM          string    `json:"president_nim"`
	ViceNIM               string    `json:"vice_nim"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}
