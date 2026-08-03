package auth

import (
	"errors"
	"fmt"
	"time"

	"christ-api/internal/auth/dto/responses"
	"christ-api/internal/auth/helpers"
	"christ-api/internal/contacts"
	"christ-api/pkg/jwt"
)

type AuthService struct {
	Repo *AuthRepository
}

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

	fmt.Printf("\n[DUMMY OTP DELIVERY] To: %s, OTP: %s (Expires in 5 minutes)\n\n", u.Email, otp)

	return otp, u, c, nil
}

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

func (s *AuthService) GoogleLoginOrRegister(email, googleID string, siteID *int64) (string, string, *responses.LoginUserResponse, error) {
	user, err := s.Repo.FindByEmail(email)
	if err != nil {
		return "", "", nil, err
	}

	if user == nil {
		user, err = s.Repo.CreateGoogleUser(email, googleID, nil, siteID, nil)
		if err != nil {
			return "", "", nil, err
		}
	}

	if user.AuthProvider != "google" {
		return "", "", nil, errors.New("this email is registered using password credentials. Please login via Email/Password")
	}

	if user.ApprovalStatus == "pending_username" {
		return "", "pending_username", &responses.LoginUserResponse{
			ID:             user.ID,
			Email:          user.Email,
			ApprovalStatus: user.ApprovalStatus,
			IsActive:       user.IsActive,
		}, nil
	}

	if user.ApprovalStatus != "approved" || !user.IsActive {
		return "", user.ApprovalStatus, &responses.LoginUserResponse{
			ID:             user.ID,
			Email:          user.Email,
			ApprovalStatus: user.ApprovalStatus,
			IsActive:       user.IsActive,
		}, nil
	}

	if err := s.Repo.UpdateLastLoginAndSite(user.ID, siteID); err != nil {
		return "", "", nil, err
	}

	token, err := jwt.GenerateToken(int(user.ID))
	if err != nil {
		return "", "", nil, err
	}

	profile, err := s.Repo.GetLoginUserProfile(user.ID)
	if err != nil {
		return "", "", nil, err
	}

	return token, "approved", profile, nil
}

func (s *AuthService) SubmitGoogleUsername(userID int64, username string) error {
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

	return s.Repo.UpdateUsernameAndStatus(userID, username, "pending_approval")
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
