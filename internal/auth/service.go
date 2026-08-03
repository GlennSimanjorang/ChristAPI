package auth

import (
	"errors"
	"fmt"
	"time"

	"christ-api/internal/auth/dto/responses"
	"christ-api/internal/auth/helpers"
	"christ-api/internal/contacts"
	emailsvc "christ-api/internal/email"
	"christ-api/pkg/jwt"
)

// AuthService struct provides methods for user authentication and management.
type AuthService struct {
	Repo *AuthRepository
}

// ya function buat login user, tapi login pake email & password, bukan google login. kalau login pake google, pakai function GoogleLoginOrRegister
func (s *AuthService) Login(email, password string, siteID *int64) (string, *responses.LoginUserResponse, error) {

	// cari user berdasarkan email
	user, err := s.Repo.FindByEmail(email)
	if err != nil {
		return "", nil, err
	}

	// kalau user tidak ditemukan, return error
	if user == nil {
		return "", nil, errors.New("user not found")
	}

	// kalau user bukan menggunakan credentials, return error
	// APA ITU CREDENTIALS? APA BEDANYA DENGAN GOOGLE LOGIN?

	// CREDENTIALS = login pake email & password
	// GOOGLE LOGIN = login pake akun google
	// jadi kalau login by google tapi usernya terdaftar pake email & password, maka user harus login pake email & password, bukan google login
	// nah kaau login by email & password tapi usernya terdaftar pake google login, maka user harus login pake google login, bukan email & password
	if user.AuthProvider != "credentials" {
		return "", nil, fmt.Errorf("please login using your %s account", user.AuthProvider)
	}

	// cek password
	if err := helpers.ComparePassword(user.Password, password); err != nil {
		return "", nil, errors.New("wrong password")
	}

	// cek status user, apakah sudah diapprove dan aktif
	if !user.IsActive || user.ApprovalStatus != "approved" {
		return "", nil, fmt.Errorf("login failed: account status is %s and inactive", user.ApprovalStatus)
	}

	// update last login dan site id
	// siteId dapet darimana? dari request body login, user bisa login ke site yang berbeda-beda, jadi siteId harus dikirim dari request body login
	if err := s.Repo.UpdateLastLoginAndSite(user.ID, siteID); err != nil {
		return "", nil, err
	}

	// generate token
	token, err := jwt.GenerateToken(int(user.ID))
	if err != nil {
		return "", nil, err
	}

	// ambil profile user untuk dikirim ke client
	profile, err := s.Repo.GetLoginUserProfile(user.ID)
	if err != nil {
		return "", nil, err
	}

	return token, profile, nil
}

// Function RegisterWithContact itu untuk register user baru sekaligus membuat contact baru, jadi nanti user baru ini akan punya contact baru juga.
func (s *AuthService) RegisterWithContact(fullName string, phone *string, address *string, contactSiteID *int64, email, password string, roleID, userSiteID *int64) (string, *User, *contacts.Contact, error) {
	existing, err := s.Repo.FindByEmail(email)
	if err != nil {
		return "", nil, nil, err
	}
	if existing != nil {
		return "", nil, nil, errors.New("user already exists")
	}

	hashed, err := helpers.HashPassword(password)
	if err != nil {
		return "", nil, nil, err
	}

	c, u, err := s.Repo.CreateContactAndUser(fullName, phone, address, contactSiteID, email, hashed, roleID, userSiteID)
	if err != nil {
		return "", nil, nil, err
	}

	otp := helpers.GenerateOTP()
	expiry := time.Now().Add(5 * time.Minute)
	if err := s.Repo.SaveOTP(u.ID, otp, expiry); err != nil {
		return "", nil, nil, err
	}

	// Send OTP via email
	if err := emailsvc.SendOTP(u.Email, otp); err != nil {
		// Log error but don't fail registration — user can still verify with console OTP if email fails
		fmt.Printf("[WARN] Failed to send OTP email to %s: %v\n", u.Email, err)
		fmt.Printf("[FALLBACK] OTP for testing: %s (expires in 5 minutes)\n", otp)
	}

	return otp, u, c, nil
}

// VerifyOTP itu fungsinya untuk verifikasi OTP yang dikirim ke email user saat register. Kalau OTP valid, maka status user akan diubah menjadi pending_approval, jadi nanti admin bisa approve user ini.
func (s *AuthService) VerifyOTP(email, otpCode string) error {
	user, err := s.Repo.FindByEmail(email)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	if user.ApprovalStatus != "pending_otp" {
		return fmt.Errorf("user is not in pending_otp status (current status: %s)", user.ApprovalStatus)
	}

	valid, err := s.Repo.VerifyOTP(user.ID, otpCode)
	if err != nil {
		return err
	}
	if !valid {
		return errors.New("invalid or expired OTP")
	}

	// Update user status to pending_approval
	return s.Repo.UpdateStatus(user.ID, "pending_approval", false)
}

// Function ini fungsinya untuk login atau register user via Google. Kalau user sudah ada, maka akan login. Kalau belum ada, maka akan register user baru dengan status pending_username, jadi nanti user harus submit username dulu sebelum bisa login.
func (s *AuthService) GoogleLoginOrRegister(email, googleID string, siteID *int64) (string, string, *responses.LoginUserResponse, error) {

	// cari user berdasarkan email
	user, err := s.Repo.FindByEmail(email)
	if err != nil {
		return "", "", nil, err
	}

	// kalau user belum ada, maka register user baru dengan status pending_username
	if user == nil {
		user, err = s.Repo.CreateGoogleUser(email, googleID, nil, siteID, nil)
		if err != nil {
			return "", "", nil, err
		}
	}

	// kalau user ada tapi bukan google, maka return error
	if user.AuthProvider != "google" {
		return "", "", nil, errors.New("this email is registered using password credentials. Please login via Email/Password")
	}

	// kalau user ada tapi statusnya pending_username, maka return status pending_username
	// fungsinya untuk memberitahu client bahwa user harus submit username dulu sebelum bisa login
	if user.ApprovalStatus == "pending_username" {
		return "", "pending_username", &responses.LoginUserResponse{
			ID:             user.ID,
			Email:          user.Email,
			ApprovalStatus: user.ApprovalStatus,
			IsActive:       user.IsActive,
		}, nil
	}

	// kalau user ada tapi statusnya pending_approval atau rejected, maka return statusnya sesuai dengan status user
	if user.ApprovalStatus == "pending_approval" || user.ApprovalStatus == "rejected" {
		return "", user.ApprovalStatus, &responses.LoginUserResponse{
			ID:             user.ID,
			Email:          user.Email,
			ApprovalStatus: user.ApprovalStatus,
			IsActive:       user.IsActive,
		}, nil
	}

	// kalau user ada dan statusnya approved, maka generate token dan return profile
	if user.ApprovalStatus != "approved" || !user.IsActive {
		return "", user.ApprovalStatus, &responses.LoginUserResponse{
			ID:             user.ID,
			Email:          user.Email,
			ApprovalStatus: user.ApprovalStatus,
			IsActive:       user.IsActive,
		}, nil
	}

	// kalau user ada dan statusnya approved, maka update last login dan site id
	if err := s.Repo.UpdateLastLoginAndSite(user.ID, siteID); err != nil {
		return "", "", nil, err
	}

	// generate token
	token, err := jwt.GenerateToken(int(user.ID))
	if err != nil {
		return "", "", nil, err
	}

	// ambil profile user untuk dikirim ke client
	profile, err := s.Repo.GetLoginUserProfile(user.ID)
	if err != nil {
		return "", "", nil, err
	}

	return token, "approved", profile, nil
}

func (s *AuthService) SubmitGoogleUsername(userID int64, username, fullName string, phone, address *string, siteID *int64) error {
	user, err := s.Repo.FindByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	if user.AuthProvider != "google" {
		return errors.New("only google users can submit username via this endpoint")
	}

	if user.ApprovalStatus != "pending_username" {
		return fmt.Errorf("user is not in pending_username status (current status: %s)", user.ApprovalStatus)
	}

	// Create contact if user doesn't have one yet
	var contactID *int64
	if user.ContactID == nil {
		contact, err := s.Repo.CreateContact(fullName, phone, address, siteID)
		if err != nil {
			return err
		}
		contactID = &contact.ID
	} else {
		contactID = user.ContactID
	}

	// Update username, contact_id, and approval status
	return s.Repo.UpdateUsernameContactAndStatus(userID, username, contactID, "pending_approval")
}

func (s *AuthService) ApproveUser(userID int64) error {
	user, err := s.Repo.FindByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	if user.ApprovalStatus != "pending_approval" {
		return fmt.Errorf("user status is %s, cannot approve unless pending_approval", user.ApprovalStatus)
	}

	return s.Repo.UpdateStatus(userID, "approved", true)
}

func (s *AuthService) RejectUser(userID int64) error {
	user, err := s.Repo.FindByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	return s.Repo.UpdateStatus(userID, "rejected", false)
}

func (s *AuthService) GetPendingApprovals() ([]User, error) {
	return s.Repo.GetPendingApprovals()
}

func (s *AuthService) Logout(userID int64) error {
	user, err := s.Repo.FindByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	return s.Repo.UpdateLastLogout(userID)
}
