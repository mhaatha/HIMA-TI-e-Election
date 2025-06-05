package helper

import (
	"github.com/mhaatha/HIMA-TI-e-Election/model/domain"
	"github.com/mhaatha/HIMA-TI-e-Election/model/web"
)

func ToLoginResponse(user domain.User) web.LoginResponse {
	return web.LoginResponse{
		ID:           user.Id,
		NIM:          user.NIM,
		FullName:     user.FullName,
		StudyProgram: user.StudyProgram,
		Role:         user.Role,
		PhoneNumber:  user.PhoneNumber,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}
