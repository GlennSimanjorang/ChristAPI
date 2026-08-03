package auth

import "time"

type User struct {
	ID             int64      `json:"id"`
	UUID           string     `json:"uuid"`
	Email          string     `json:"email"`
	Username       *string    `json:"username"`
	Password       string     `json:"-"`
	GoogleID       *string    `json:"google_id,omitempty"`
	AuthProvider   string     `json:"auth_provider"`   // "credentials" or "google"
	ApprovalStatus string     `json:"approval_status"` // "pending_username", "pending_otp", "pending_approval", "approved", "rejected"
	PointsBalance  int64      `json:"points_balance"`
	RoleID         *int64     `json:"role_id"`
	ContactID      *int64     `json:"contact_id"`
	IsActive       bool       `json:"is_active"`
	LastLoginAt    *time.Time `json:"last_login_at"`
	CreatedAt      *time.Time `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at"`
	SiteID         *int64     `json:"site_id"`
}

type UserOTP struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	OTPCode   string    `json:"otp_code"`
	ExpiredAt time.Time `json:"expired_at"`
	CreatedAt time.Time `json:"created_at"`
}
