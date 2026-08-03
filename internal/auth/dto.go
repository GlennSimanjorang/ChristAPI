package auth

type LoginResponse struct {
	Token string  `json:"token"`
	User  UserDTO `json:"user"`
}

type LoginUserResponse struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	Email          string  `json:"email"`
	Username       *string `json:"username"`
	Role           string  `json:"role"`
	Points         int64   `json:"points"`
	ApprovalStatus string  `json:"approval_status"`
	IsActive       bool    `json:"is_active"`
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

type UserDTO struct {
	ID             int64   `json:"id"`
	Email          string  `json:"email"`
	Username       *string `json:"username"`
	ApprovalStatus string  `json:"approval_status"`
	IsActive       bool    `json:"is_active"`
}

type GoogleLoginRequest struct {
	IDToken string `json:"id_token"`
	Email   string `json:"email"` // fallback/mock
}

type SubmitUsernameRequest struct {
	Username string `json:"username"`
}

type VerifyOTPRequest struct {
	Email   string `json:"email"`
	OTPCode string `json:"otp_code"`
}
