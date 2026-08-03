package role

type RoleService struct {
	Repo *RoleRepository
}

func (s *RoleService) List(id, siteID *int64) ([]Role, error) {
	return s.Repo.Get(id, siteID)
}

func (s *RoleService) Create(name string, code string, description *string, siteID *int64) (*Role, error) {
	return s.Repo.Create(name, code, description, siteID)
}

func (s *RoleService) Update(id int64, name string, code string, description *string) (*Role, error) {
	return s.Repo.Update(id, name, code, description)
}

func (s *RoleService) GetByID(id int64) (*Role, error) {
	return s.Repo.GetByID(id)
}

func (s *RoleService) HasPermission(roleID int64, requiredCode string) (bool, error) {
	role, err := s.Repo.GetByID(roleID)
	if err != nil {
		return false, err
	}
	return role.Code == requiredCode, nil
}
