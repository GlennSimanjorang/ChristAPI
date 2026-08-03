package role

import (
	"strconv"

	"christ-api/internal/role/dto/requests"
	"christ-api/internal/role/helpers"
	"christ-api/pkg/response"

	"github.com/gofiber/fiber/v2"
)

var Service = RoleService{}

func InitService(repo *RoleRepository) {
	if repo != nil {
		Service = RoleService{Repo: repo}
	}
}

func ListRoles(c *fiber.Ctx) error {
	req := &requests.ListRolesRequest{}

	// check request query parameters
	if err := c.QueryParser(req); err != nil {
		return response.Error(c, 422, "Invalid query parameters", nil)
	}

	// validasiin request query parameters
	roles, err := Service.List(req.ID, req.SiteID)
	if err != nil {
		return response.Error(c, 500, "Failed to list roles", nil)
	}

	// convert roles to response DTOs
	resp := RolesToRoleResponses(roles)
	return response.Success(c, "Roles retrieved", resp)
}

func CreateRole(c *fiber.Ctx) error {
	req := new(requests.CreateRoleRequest)
	if err := c.BodyParser(req); err != nil {
		return response.Error(c, 422, "Invalid request", nil)
	}

	if err := helpers.ValidateCreateRoleRequest(req); err != nil {
		return response.Error(c, 422, err.Error(), nil)
	}

	role, err := Service.Create(req.Name, req.Code, req.Description, req.SiteID)
	if err != nil {
		return response.Error(c, 500, "Failed to create role", nil)
	}

	return response.Created(c, "Role created", RoleToRoleResponse(role))
}

func UpdateRole(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return response.Error(c, 400, "Invalid role ID", nil)
	}

	req := new(requests.UpdateRoleRequest)
	if err := c.BodyParser(req); err != nil {
		return response.Error(c, 422, "Invalid request", nil)
	}

	if err := helpers.ValidateUpdateRoleRequest(req); err != nil {
		return response.Error(c, 422, err.Error(), nil)
	}

	role, err := Service.Update(id, req.Name, req.Code, req.Description)
	if err != nil {
		return response.Error(c, 500, "Failed to update role", nil)
	}

	return response.Success(c, "Role updated", RoleToRoleResponse(role))
}
