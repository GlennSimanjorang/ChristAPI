package auth

import (
	"christ-api/internal/auth/dto/responses"
)

// UserToUserDTO converts User model to UserDTO
func UserToUserDTO(user *User) responses.UserDTO {
	return responses.UserDTO{
		ID:             user.ID,
		Email:          user.Email,
		Username:       user.Username,
		ApprovalStatus: user.ApprovalStatus,
		IsActive:       user.IsActive,
	}
}

// UsersToUserDTOs converts multiple User models to UserDTOs
func UsersToUserDTOs(users []User) []responses.UserDTO {
	var dtos []responses.UserDTO
	for _, u := range users {
		dtos = append(dtos, UserToUserDTO(&u))
	}
	return dtos
}

// LoginUserResponseToLoginDataResponse creates LoginDataResponse from LoginUserResponse and token
func LoginUserResponseToLoginDataResponse(profile *responses.LoginUserResponse, token string) responses.LoginDataResponse {
	return responses.LoginDataResponse{
		User:  *profile,
		Token: token,
	}
}
