package responses

type RoleResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	SiteID      *int64  `json:"site_id"`
}
