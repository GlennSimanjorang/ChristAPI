package responses

type UserDTO struct {
	ID             int64   `json:"id"`
	Email          string  `json:"email"`
	Username       *string `json:"username"`
	ApprovalStatus string  `json:"approval_status"`
	IsActive       bool    `json:"is_active"`
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
