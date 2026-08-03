package responses

type LoginResponse struct {
	Token string  `json:"token"`
	User  UserDTO `json:"user"`
}

type LoginDataResponse struct {
	User  LoginUserResponse `json:"user"`
	Token string            `json:"token"`
}

type LoginSuccessResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Data    LoginDataResponse `json:"data"`
}
