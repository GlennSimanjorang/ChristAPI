package auth

import (
	"database/sql"
	"errors"
	"time"

	"christ-api/internal/contacts"
)

type AuthRepository struct {
	DB *sql.DB
}

func (r *AuthRepository) scanUser(row scanner) (*User, error) {
	var user User
	var roleID sql.NullInt64
	var contactID sql.NullInt64
	var siteID sql.NullInt64
	var lastLogin sql.NullTime
	var createdAt sql.NullTime
	var updatedAt sql.NullTime
	var passwordHash sql.NullString
	var username sql.NullString
	var googleID sql.NullString

	err := row.Scan(
		&user.ID, &user.UUID, &user.Email, &username, &passwordHash,
		&googleID, &user.AuthProvider, &user.ApprovalStatus,
		&roleID, &contactID, &user.IsActive, &lastLogin,
		&createdAt, &updatedAt, &siteID, &user.PointsBalance,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if passwordHash.Valid {
		user.Password = passwordHash.String
	}
	if username.Valid {
		v := username.String
		user.Username = &v
	}
	if googleID.Valid {
		v := googleID.String
		user.GoogleID = &v
	}
	if roleID.Valid {
		v := roleID.Int64
		user.RoleID = &v
	}
	if contactID.Valid {
		v := contactID.Int64
		user.ContactID = &v
	}
	if siteID.Valid {
		v := siteID.Int64
		user.SiteID = &v
	}
	if lastLogin.Valid {
		user.LastLoginAt = &lastLogin.Time
	}
	if createdAt.Valid {
		user.CreatedAt = &createdAt.Time
	}
	if updatedAt.Valid {
		user.UpdatedAt = &updatedAt.Time
	}

	return &user, nil
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func (r *AuthRepository) FindByEmail(email string) (*User, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}

	query := `SELECT id, uuid, email, username, password_hash, google_id, auth_provider, approval_status, role_id, contact_id, is_active, last_login_at, created_at, updated_at, site_id, points_balance FROM users WHERE email = $1 LIMIT 1`
	return r.scanUser(r.DB.QueryRow(query, email))
}

func (r *AuthRepository) FindByID(id int64) (*User, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}

	query := `SELECT id, uuid, email, username, password_hash, google_id, auth_provider, approval_status, role_id, contact_id, is_active, last_login_at, created_at, updated_at, site_id, points_balance FROM users WHERE id = $1 LIMIT 1`
	return r.scanUser(r.DB.QueryRow(query, id))
}

func (r *AuthRepository) CreateUser(email, passwordHash string, roleID, siteID, contactID *int64) (*User, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}

	query := `INSERT INTO users (email, password_hash, role_id, site_id, contact_id, is_active, auth_provider, approval_status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, FALSE, 'credentials', 'pending_otp', NOW(), NOW()) RETURNING id, uuid, email, username, password_hash, google_id, auth_provider, approval_status, role_id, contact_id, is_active, last_login_at, created_at, updated_at, site_id, points_balance`
	return r.scanUser(r.DB.QueryRow(query, email, passwordHash, roleID, siteID, contactID))
}

func (r *AuthRepository) CreateGoogleUser(email, googleID string, roleID, siteID, contactID *int64) (*User, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}

	query := `INSERT INTO users (email, google_id, role_id, site_id, contact_id, is_active, auth_provider, approval_status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, FALSE, 'google', 'pending_username', NOW(), NOW()) RETURNING id, uuid, email, username, password_hash, google_id, auth_provider, approval_status, role_id, contact_id, is_active, last_login_at, created_at, updated_at, site_id, points_balance`
	return r.scanUser(r.DB.QueryRow(query, email, googleID, roleID, siteID, contactID))
}

func (r *AuthRepository) UpdateUsernameAndStatus(userID int64, username, newStatus string) error {
	if r == nil || r.DB == nil {
		return sql.ErrConnDone
	}

	// Check if username is already taken
	var exists bool
	err := r.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND id != $2)", username, userID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("username is already taken")
	}

	query := `UPDATE users SET username = $2, approval_status = $3, updated_at = NOW() WHERE id = $1`
	_, err = r.DB.Exec(query, userID, username, newStatus)
	return err
}

func (r *AuthRepository) UpdateStatus(userID int64, newStatus string, isActive bool) error {
	if r == nil || r.DB == nil {
		return sql.ErrConnDone
	}

	query := `UPDATE users SET approval_status = $2, is_active = $3, updated_at = NOW() WHERE id = $1`
	_, err := r.DB.Exec(query, userID, newStatus, isActive)
	return err
}

func (r *AuthRepository) UpdateLastLoginAndSite(userID int64, siteID *int64) error {
	if r == nil || r.DB == nil {
		return sql.ErrConnDone
	}

	query := `UPDATE users SET last_login_at = NOW(), site_id = COALESCE($2, site_id) WHERE id = $1`
	_, err := r.DB.Exec(query, userID, siteID)
	return err
}

func (r *AuthRepository) GetLoginUserProfile(userID int64) (*LoginUserResponse, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}

	query := `
		SELECT
			u.id,
			COALESCE(c.full_name, ''),
			u.email,
			u.username,
			COALESCE(ro.name, ''),
			COALESCE(u.points_balance, 0),
			u.approval_status,
			u.is_active
		FROM users u
		LEFT JOIN contacts c ON c.id = u.contact_id
		LEFT JOIN roles ro ON ro.id = u.role_id
		WHERE u.id = $1
		LIMIT 1`

	var p LoginUserResponse
	var username sql.NullString
	row := r.DB.QueryRow(query, userID)
	if err := row.Scan(&p.ID, &p.Name, &p.Email, &username, &p.Role, &p.Points, &p.ApprovalStatus, &p.IsActive); err != nil {
		return nil, err
	}
	if username.Valid {
		v := username.String
		p.Username = &v
	}

	return &p, nil
}

func (r *AuthRepository) CreateContactAndUser(fullName string, phone *string, address *string, contactSiteID *int64, email, passwordHash string, roleID, userSiteID *int64) (*contacts.Contact, *User, error) {
	if r == nil || r.DB == nil {
		return nil, nil, sql.ErrConnDone
	}

	tx, err := r.DB.Begin()
	if err != nil {
		return nil, nil, err
	}
	rollback := func() { _ = tx.Rollback() }

	// insert contact
	var c contacts.Contact
	contactQuery := `INSERT INTO contacts (full_name, phone, address, created_at, updated_at, site_id) VALUES ($1,$2,$3,NOW(),NOW(),$4) RETURNING id, full_name, phone, address, created_at, updated_at, site_id`
	var phoneN sql.NullString
	var addrN sql.NullString
	var createdN sql.NullTime
	var updatedN sql.NullTime
	var siteIDN sql.NullInt64

	row := tx.QueryRow(contactQuery, fullName, phone, address, contactSiteID)
	if err := row.Scan(&c.ID, &c.FullName, &phoneN, &addrN, &createdN, &updatedN, &siteIDN); err != nil {
		rollback()
		return nil, nil, err
	}
	if phoneN.Valid {
		v := phoneN.String
		c.Phone = &v
	}
	if addrN.Valid {
		v := addrN.String
		c.Address = &v
	}
	if createdN.Valid {
		c.CreatedAt = &createdN.Time
	}
	if updatedN.Valid {
		c.UpdatedAt = &updatedN.Time
	}
	if siteIDN.Valid {
		v := siteIDN.Int64
		c.SiteID = &v
	}

	// insert user with contact_id
	userQuery := `INSERT INTO users (email, password_hash, role_id, site_id, contact_id, is_active, auth_provider, approval_status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,FALSE,'credentials','pending_otp',NOW(),NOW()) RETURNING id, uuid, email, username, password_hash, google_id, auth_provider, approval_status, role_id, contact_id, is_active, last_login_at, created_at, updated_at, site_id, points_balance`
	u, err := r.scanUser(tx.QueryRow(userQuery, email, passwordHash, roleID, userSiteID, c.ID))
	if err != nil {
		rollback()
		return nil, nil, err
	}

	if err := tx.Commit(); err != nil {
		rollback()
		return nil, nil, err
	}

	return &c, u, nil
}

// SaveOTP saves an OTP to database
func (r *AuthRepository) SaveOTP(userID int64, otpCode string, expiry time.Time) error {
	if r == nil || r.DB == nil {
		return sql.ErrConnDone
	}

	// Delete old OTPs first
	_, _ = r.DB.Exec("DELETE FROM user_otps WHERE user_id = $1", userID)

	query := `INSERT INTO user_otps (user_id, otp_code, expired_at) VALUES ($1, $2, $3)`
	_, err := r.DB.Exec(query, userID, otpCode, expiry)
	return err
}

// VerifyOTP verifies if the OTP is correct and not expired, then deletes it
func (r *AuthRepository) VerifyOTP(userID int64, otpCode string) (bool, error) {
	if r == nil || r.DB == nil {
		return false, sql.ErrConnDone
	}

	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM user_otps WHERE user_id = $1 AND otp_code = $2 AND expired_at > NOW())`
	err := r.DB.QueryRow(query, userID, otpCode).Scan(&exists)
	if err != nil {
		return false, err
	}

	if exists {
		// Clean up the OTP
		_, _ = r.DB.Exec("DELETE FROM user_otps WHERE user_id = $1", userID)
		return true, nil
	}

	return false, nil
}

// GetPendingApprovals returns a list of users pending admin approval
func (r *AuthRepository) GetPendingApprovals() ([]User, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}

	query := `SELECT id, uuid, email, username, password_hash, google_id, auth_provider, approval_status, role_id, contact_id, is_active, last_login_at, created_at, updated_at, site_id, points_balance FROM users WHERE approval_status = 'pending_approval' ORDER BY created_at ASC`
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		u, err := r.scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *u)
	}

	return users, nil
}
