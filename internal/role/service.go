package role

type RoleService struct {
	Repo *RoleRepository
}

func (s *RoleService) List(id, siteID *int64) ([]Role, error) {
	return s.Repo.Get(id, siteID)
}

func (s *RoleService) Create(name string, description *string, siteID *int64) (*Role, error) {
	return s.Repo.Create(name, description, siteID)
}

func (s *RoleService) Update(id int64, name string, description *string) (*Role, error) {
	return s.Repo.Update(id, name, description)
}

