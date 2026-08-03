package role

import (
	"database/sql"
)

type RoleRepository struct {
	DB *sql.DB
}

func (r *RoleRepository) Get(id, siteID *int64) ([]Role, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}

	var rows *sql.Rows
	var err error

	// Query based on the provided parameters
	// kalau semisal ga ngasih request id maka ambil semua role yang ada di database
	if id != nil {
		rows, err = r.DB.Query(`SELECT id, name, code, description, site_id FROM roles WHERE id = $1`, *id)
	
	// nah kalau ngasih request site_id maka ambil semua role yang ada di database berdasarkan site_id
	} else if siteID != nil {
		rows, err = r.DB.Query(`SELECT id, name, code, description, site_id FROM roles WHERE site_id = $1`, *siteID)
	// kalau ga ngasih request id dan site_id maka ambil semua role yang ada di database
	} else {
		rows, err = r.DB.Query(`SELECT id, name, code, description, site_id FROM roles`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Role
	for rows.Next() {
		var role Role
		var desc sql.NullString
		var siteID sql.NullInt64
		if err := rows.Scan(
			&role.ID, 
			&role.Name, 
			&role.Code, 
			&desc, 
			&siteID); err != nil {
			return nil, err
		}
		
		if desc.Valid {
			v := desc.String
			role.Description = &v
		}
		if siteID.Valid {
			v := siteID.Int64
			role.SiteID = &v
		}
		out = append(out, role)
	}
	return out, nil
}

func (r *RoleRepository) Create(name string, code string, description *string, siteID *int64) (*Role, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}

	query := `INSERT INTO roles (name, code, description, site_id) VALUES ($1, $2, $3, $4) RETURNING id, name, code, description, site_id`
	var role Role
	var desc sql.NullString
	var sID sql.NullInt64
	row := r.DB.QueryRow(query, name, code, description, siteID)
	if err := row.Scan(&role.ID, &role.Name, &role.Code, &desc, &sID); err != nil {
		return nil, err
	}
	if desc.Valid {
		v := desc.String
		role.Description = &v
	}
	if sID.Valid {
		v := sID.Int64
		role.SiteID = &v
	}
	return &role, nil
}

func (r *RoleRepository) Update(id int64, name string, code string, description *string) (*Role, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}

	query := `UPDATE roles SET name = $1, code = $2, description = $3 WHERE id = $4 RETURNING id, name, code, description, site_id`
	var role Role
	var desc sql.NullString
	var sID sql.NullInt64
	row := r.DB.QueryRow(query, name, code, description, id)
	if err := row.Scan(&role.ID, &role.Name, &role.Code, &desc, &sID); err != nil {
		return nil, err
	}
	if desc.Valid {
		v := desc.String
		role.Description = &v
	}
	if sID.Valid {
		v := sID.Int64
		role.SiteID = &v
	}
	return &role, nil
}

func (r *RoleRepository) GetByID(id int64) (*Role, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}

	query := `SELECT id, name, code, description, site_id FROM roles WHERE id = $1`
	var role Role
	var desc sql.NullString
	var sID sql.NullInt64
	row := r.DB.QueryRow(query, id)
	if err := row.Scan(&role.ID, &role.Name, &role.Code, &desc, &sID); err != nil {
		return nil, err
	}
	if desc.Valid {
		v := desc.String
		role.Description = &v
	}
	if sID.Valid {
		v := sID.Int64
		role.SiteID = &v
	}
	return &role, nil
}
