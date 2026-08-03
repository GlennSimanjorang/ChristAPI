package requests

type ListRolesRequest struct {
	ID     *int64 `query:"id"`
	SiteID *int64 `query:"siteId"`
}

type CreateRoleRequest struct {
	Name        string  `json:"name"`
	Code        string  `json:"code"`
	Description *string `json:"description"`
	SiteID      *int64  `json:"site_id"`
}

type UpdateRoleRequest struct {
	Name        string  `json:"name"`
	Code        string  `json:"code"`
	Description *string `json:"description"`
}
