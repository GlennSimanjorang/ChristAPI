package auth

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"christ-api/internal/contacts"
	"christ-api/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	Repo *AuthRepository
}

func (s *AuthService) Login(email, password string, siteID *int64) (string, *LoginUserResponse, error) {
	user, err := s.Repo.FindByEmail(email)
	if err != nil {
		return "", nil, err
	}
	if user == nil {
		return "", nil, errors.New("user not found")
	}

	if user.AuthProvider != "credentials" {
		return "", nil, fmt.Errorf("please login using your %s account", user.AuthProvider)
	}

	// compare hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, errors.New("wrong password")
	}

	// Check user state
	if !user.IsActive || user.ApprovalStatus != "approved" {
		return "", nil, fmt.Errorf("login failed: account status is %s and inactive", user.ApprovalStatus)
	}

	// update last login
	if err := s.Repo.UpdateLastLoginAndSite(user.ID, siteID); err != nil {
		return "", nil, err
	}

	token, err := jwt.GenerateToken(int(user.ID))
	if err != nil {
		return "", nil, err
	}

	profile, err := s.Repo.GetLoginUserProfile(user.ID)
	if err != nil {
		return "", nil, err
	}

	return token, profile, nil
}

// GenerateOTP generates a random 6-digit number
func GenerateOTP() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return fmt.Sprintf("%06d", n.Int64()+100000)
}

// RegisterWithContact creates a contact and user in a single transaction and generates a dummy OTP.
func (s *AuthService) RegisterWithContact(fullName string, phone *string, address *string, contactSiteID *int64, email, password string, roleID, userSiteID *int64) (string, *User, *contacts.Contact, error) {
	// check existing user
	existing, err := s.Repo.FindByEmail(email)
	if err != nil {
		return "", nil, nil, err
	}
	if existing != nil {
		return "", nil, nil, errors.New("user already exists")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, nil, err
	}

	c, u, err := s.Repo.CreateContactAndUser(fullName, phone, address, contactSiteID, email, string(hashed), roleID, userSiteID)
	if err != nil {
		return "", nil, nil, err
	}

	// Generate OTP
	otp := GenerateOTP()
	expiry := time.Now().Add(5 * time.Minute)
	if err := s.Repo.SaveOTP(u.ID, otp, expiry); err != nil {
		return "", nil, nil, err
	}

	// We print the OTP as a dummy delivery helper in dev mode
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

func (s *AuthService) GoogleLoginOrRegister(email, googleID string, siteID *int64) (string, string, *LoginUserResponse, error) {
	user, err := s.Repo.FindByEmail(email)
	if err != nil {
		return "", "", nil, err
	}

	// If user does not exist, register them
	if user == nil {
		user, err = s.Repo.CreateGoogleUser(email, googleID, nil, siteID, nil)
		if err != nil {
			return "", "", nil, err
		}
	}

	if user.AuthProvider != "google" {
		return "", "", nil, errors.New("this email is registered using password credentials. Please login via Email/Password")
	}

	// Check status
	if user.ApprovalStatus == "pending_username" {
		// Needs to submit username next
		return "", "pending_username", &LoginUserResponse{
			ID:             user.ID,
			Email:          user.Email,
			ApprovalStatus: user.ApprovalStatus,
			IsActive:       user.IsActive,
		}, nil
	}

	if user.ApprovalStatus != "approved" || !user.IsActive {
		return "", user.ApprovalStatus, &LoginUserResponse{
			ID:             user.ID,
			Email:          user.Email,
			ApprovalStatus: user.ApprovalStatus,
			IsActive:       user.IsActive,
		}, nil
	}

	// update last login
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
