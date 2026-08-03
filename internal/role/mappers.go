package role

import (
	"christ-api/internal/role/dto/responses"
)

func RoleToRoleResponse(role *Role) *responses.RoleResponse {
	if role == nil {
		return nil
	}
	return &responses.RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Code:        role.Code,
		Description: role.Description,
		SiteID:      role.SiteID,
	}
}

func RolesToRoleResponses(roles []Role) []responses.RoleResponse {
	if len(roles) == 0 {
		return []responses.RoleResponse{}
	}
	resp := make([]responses.RoleResponse, len(roles))
	for i, r := range roles {
		resp[i] = *RoleToRoleResponse(&r)
	}
	return resp
}
