package web

type CandidateCreateRequest struct {
	Number                int      `json:"number" validate:"required,min=1"`
	President             string   `json:"president" validate:"required,min=3,max=255"`
	Vice                  string   `json:"vice" validate:"required,min=3,max=255"`
	Vision                string   `json:"vision" validate:"omitempty,min=3"`
	Mission               []string `json:"mission" validate:"omitempty,min=1"`
	PhotoKey              string   `json:"photo_key" validate:"required"`
	PresidentStudyProgram string   `json:"president_study_program" validate:"required,min=3,max=100"`
	ViceStudyProgram      string   `json:"vice_study_program" validate:"required,min=3,max=100"`
	PresidentNIM          string   `json:"president_nim" validate:"required,min=4,max=14"`
	ViceNIM               string   `json:"vice_nim" validate:"required,min=4,max=14"`
}

type CandidateUpdateRequest struct {
	Number                int      `json:"number" validate:"omitempty,min=1"`
	President             string   `json:"president" validate:"omitempty,min=3,max=255"`
	Vice                  string   `json:"vice" validate:"omitempty,min=3,max=255"`
	Vision                string   `json:"vision" validate:"omitempty,min=3"`
	Mission               []string `json:"mission" validate:"omitempty,min=1"`
	PhotoKey              string   `json:"photo_key" validate:"omitempty"`
	PresidentStudyProgram string   `json:"president_study_program" validate:"omitempty,min=3,max=100"`
	ViceStudyProgram      string   `json:"vice_study_program" validate:"omitempty,min=3,max=100"`
	PresidentNIM          string   `json:"president_nim" validate:"omitempty,min=4,max=14"`
	ViceNIM               string   `json:"vice_nim" validate:"omitempty,min=4,max=14"`
}
