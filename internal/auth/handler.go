package auth

import (
	"context"
	"log"
	"os"

	"christ-api/internal/contacts"
	"christ-api/pkg/response"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/api/idtoken"
)

var service = AuthService{}

// InitService initializes package-level auth service with a repository.
func InitService(repo *AuthRepository) {
	if repo != nil {
		service = AuthService{Repo: repo}
	}
}

func Login(c *fiber.Ctx) error {
	type Request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		SiteID   *int64 `json:"site_id"`
	}

	req := new(Request)

	if err := c.BodyParser(req); err != nil {
		return response.Error(c, 422, "Invalid request", nil)
	}

	token, profile, err := service.Login(req.Email, req.Password, req.SiteID)
	if err != nil {
		return response.Error(c, 401, err.Error(), nil)
	}

	data := LoginDataResponse{User: *profile, Token: token}
	return response.Success(c, "Login berhasil", data)
}

func Register(c *fiber.Ctx) error {
	type Request struct {
		FullName      string  `json:"full_name"`
		Phone         *string `json:"phone"`
		Address       *string `json:"address"`
		ContactSiteID *int64  `json:"contact_site_id"`
		Email         string  `json:"email"`
		Password      string  `json:"password"`
		RoleID        *int64  `json:"role_id"`
		SiteID        *int64  `json:"site_id"`
	}

	req := new(Request)
	if err := c.BodyParser(req); err != nil {
		return response.Error(c, 422, "Invalid request", nil)
	}

	validationErrs := make(map[string][]string)
	if req.FullName == "" {
		validationErrs["full_name"] = append(validationErrs["full_name"], "Full name is required")
	}
	if req.Email == "" {
		validationErrs["email"] = append(validationErrs["email"], "Email is required")
	}
	if req.Password == "" {
		validationErrs["password"] = append(validationErrs["password"], "Password is required")
	}
	if len(validationErrs) > 0 {
		return response.Error(c, 422, "Validation failed", validationErrs)
	}

	otp, user, contact, err := service.RegisterWithContact(req.FullName, req.Phone, req.Address, req.ContactSiteID, req.Email, req.Password, req.RoleID, req.SiteID)
	if err != nil {
		return response.Error(c, 500, err.Error(), nil)
	}

	resp := struct {
		User    UserDTO           `json:"user"`
		Contact *contacts.Contact `json:"contact"`
		OTP     string            `json:"otp"` // Returned in dev/dummy mode so frontend can read it
	}{
		User:    UserDTO{ID: user.ID, Email: user.Email, ApprovalStatus: user.ApprovalStatus, IsActive: user.IsActive},
		Contact: contact,
		OTP:     otp,
	}

	return response.Created(c, "User registered. Please verify using OTP sent.", resp)
}

func VerifyOTP(c *fiber.Ctx) error {
	req := new(VerifyOTPRequest)
	if err := c.BodyParser(req); err != nil {
		return response.Error(c, 422, "Invalid request", nil)
	}

	if req.Email == "" || req.OTPCode == "" {
		return response.Error(c, 422, "Email and OTP code are required", nil)
	}

	err := service.VerifyOTP(req.Email, req.OTPCode)
	if err != nil {
		return response.Error(c, 400, err.Error(), nil)
	}

	return response.Success(c, "OTP verified successfully. Your account is now pending admin approval.", nil)
}

func LoginGoogle(c *fiber.Ctx) error {
	req := new(GoogleLoginRequest)
	if err := c.BodyParser(req); err != nil {
		return response.Error(c, 422, "Invalid request", nil)
	}

	if req.IDToken == "" {
		return response.Error(c, 422, "ID token is required", nil)
	}

	// Verify Google ID token
	payload, err := idtoken.Validate(context.Background(), req.IDToken, os.Getenv("GOOGLE_CLIENT_ID"))
	if err != nil {
		log.Printf("❌ Google token verification failed: %v", err)
		return response.Error(c, 401, "Invalid Google token", nil)
	}

	email, _ := payload.Claims["email"].(string)
	googleID, _ := payload.Claims["sub"].(string)

	if email == "" || googleID == "" {
		return response.Error(c, 401, "Missing email or subject in token", nil)
	}

	log.Printf("✅ Google OAuth verified: %s (sub: %s)", email, googleID)

	token, status, profile, err := service.GoogleLoginOrRegister(email, googleID, nil)
	if err != nil {
		return response.Error(c, 400, err.Error(), nil)
	}

	if status == "pending_username" {
		return response.Success(c, "Google registration successful. Please choose a username.", fiber.Map{
			"status": "pending_username",
			"user":   profile,
		})
	}

	if status != "approved" {
		return response.Success(c, "Google login successful but status is pending/rejected.", fiber.Map{
			"status": status,
			"user":   profile,
		})
	}

	data := LoginDataResponse{User: *profile, Token: token}
	return response.Success(c, "Google Login berhasil", data)
}

func SubmitGoogleUsername(c *fiber.Ctx) error {
	type Request struct {
		UserID   int64  `json:"user_id"`
		Username string `json:"username"`
	}

	req := new(Request)
	if err := c.BodyParser(req); err != nil {
		return response.Error(c, 422, "Invalid request", nil)
	}

	if req.UserID == 0 || req.Username == "" {
		return response.Error(c, 422, "User ID and Username are required", nil)
	}

	err := service.SubmitGoogleUsername(req.UserID, req.Username)
	if err != nil {
		return response.Error(c, 400, err.Error(), nil)
	}

	return response.Success(c, "Username updated. Your account is now pending admin approval.", nil)
}

// GetPendingApprovals list users waiting for admin approval
func GetPendingApprovals(c *fiber.Ctx) error {
	users, err := service.GetPendingApprovals()
	if err != nil {
		return response.Error(c, 500, err.Error(), nil)
	}

	var dtos []UserDTO
	for _, u := range users {
		dtos = append(dtos, UserDTO{
			ID:             u.ID,
			Email:          u.Email,
			Username:       u.Username,
			ApprovalStatus: u.ApprovalStatus,
			IsActive:       u.IsActive,
		})
	}

	return response.Success(c, "List pending approvals", dtos)
}

// ApproveUser approves the user and activates them
func ApproveUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return response.Error(c, 400, "Invalid User ID", nil)
	}

	err = service.ApproveUser(int64(id))
	if err != nil {
		return response.Error(c, 400, err.Error(), nil)
	}

	return response.Success(c, "User approved successfully.", nil)
}

// RejectUser rejects the user
func RejectUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return response.Error(c, 400, "Invalid User ID", nil)
	}

	err = service.RejectUser(int64(id))
	if err != nil {
		return response.Error(c, 400, err.Error(), nil)
	}

	return response.Success(c, "User rejected successfully.", nil)
}
