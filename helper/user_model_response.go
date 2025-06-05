package helper

import (
	"github.com/mhaatha/HIMA-TI-e-Election/model/domain"
	"github.com/mhaatha/HIMA-TI-e-Election/model/web"
)

func ToUserResponse(user domain.User) web.UserResponse {
	return web.UserResponse{
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

func ToUserResponses(users []domain.User) []web.UserResponse {
	var userResponses []web.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, ToUserResponse(user))
	}
	return userResponses
}

func ToUserGetByNimResponse(user domain.User) web.UserGetByNimResponse {
	return web.UserGetByNimResponse{
		ID:           user.Id,
		NIM:          user.NIM,
		FullName:     user.FullName,
		StudyProgram: user.StudyProgram,
		Password:     user.Password,
		Role:         user.Role,
		PhoneNumber:  user.PhoneNumber,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

func ToSessionResponse(session domain.Session) web.SessionResponse {
	return web.SessionResponse{
		SessionId:     session.SessionId,
		UserId:        session.UserId,
		CreatedAt:     session.CreatedAt,
		MaxAgeSeconds: session.MaxAgeSeconds,
	}
}

func ToAdminResponse(admin domain.User) web.AdminResponse {
	return web.AdminResponse{
		ID:           admin.Id,
		NIM:          admin.NIM,
		FullName:     admin.FullName,
		StudyProgram: admin.StudyProgram,
		Password:     admin.Password,
		Role:         admin.Role,
		PhoneNumber:  admin.PhoneNumber,
		CreatedAt:    admin.CreatedAt,
		UpdatedAt:    admin.UpdatedAt,
	}
}

func ToAdminResponses(admins []domain.User) []web.AdminResponse {
	var adminResponses []web.AdminResponse
	for _, admin := range admins {
		adminResponses = append(adminResponses, ToAdminResponse(admin))
	}
	return adminResponses
}
