package helpers

import (
	"errors"

	"christ-api/internal/role/dto/requests"
)

func ValidateCreateRoleRequest(req *requests.CreateRoleRequest) error {
	if req.Name == "" {
		return errors.New("name is required")
	}
	if req.Code == "" {
		return errors.New("code is required")
	}
	return nil
}

func ValidateUpdateRoleRequest(req *requests.UpdateRoleRequest) error {
	if req.Name == "" {
		return errors.New("name is required")
	}
	if req.Code == "" {
		return errors.New("code is required")
	}
	return nil
}
