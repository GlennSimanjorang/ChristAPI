package requests

type RegisterRequest struct {
	FullName      string  `json:"full_name"`
	Phone         *string `json:"phone"`
	Address       *string `json:"address"`
	ContactSiteID *int64  `json:"contact_site_id"`
	Email         string  `json:"email"`
	Password      string  `json:"password"`
	RoleID        *int64  `json:"role_id"`
	SiteID        *int64  `json:"site_id"`
}
