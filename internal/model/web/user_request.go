package web

type UserCreateRequest struct {
	NIM          string `json:"nim" validate:"omitempty,min=4,max=14"`
	FullName     string `json:"full_name" validate:"required,min=3,max=100"`
	StudyProgram string `json:"study_program" validate:"omitempty,min=3,max=100"`
	Password     string `json:"password" validate:"omitempty,min=6,max=100"`
	Role         string `json:"role"`
	PhoneNumber  string `json:"phone_number" validate:"required,numericstr,min=8,max=14"`
}

type UserUpdateCurrentRequest struct {
	NIM          string `json:"nim" validate:"omitempty,min=4,max=14"`
	FullName     string `json:"full_name" validate:"omitempty,min=3,max=100"`
	StudyProgram string `json:"study_program" validate:"omitempty,min=3,max=100"`
}

type UserUpdateByIdRequest struct {
	NIM          string `json:"nim" validate:"omitempty,min=4,max=14"`
	FullName     string `json:"full_name" validate:"omitempty,min=3,max=100"`
	StudyProgram string `json:"study_program" validate:"omitempty,min=3,max=100"`
	Password     string `json:"password" validate:"omitempty,min=6,max=100"`
	Role         string `json:"role"`
	PhoneNumber  string `json:"phone_number" validate:"omitempty,numericstr,min=8,max=14"`
}
