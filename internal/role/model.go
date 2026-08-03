package role

type Role struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Code        string  `json:"code"`
	Description *string `json:"description"`
	SiteID      *int64  `json:"site_id"`
}
