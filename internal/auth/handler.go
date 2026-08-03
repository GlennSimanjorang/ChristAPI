package auth

import (
	"context"
	"log"
	"os"

	"christ-api/internal/auth/dto/requests"
	"christ-api/internal/auth/helpers"
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

	// buat request struct untuk login request
	req := new(requests.LoginRequest)

	// parse body request ke struct
	if err := c.BodyParser(req); err != nil {
		// kalau parsing gagal, return error 422
		return response.Error(c, 422, "Invalid request", nil)
	}

	// kalau email & password kosong, return error 422
	if err := helpers.ValidateLoginRequest(req); err != nil {
		return response.Error(c, 422, err.Error(), nil)
	}

	// panggil service untuk login
	token, profile, err := service.Login(req.Email, req.Password, req.SiteID)

	// kalau login gagal, return error 401
	if err != nil {
		return response.Error(c, 401, err.Error(), nil)
	}

	// kalau login berhasil, return token & profile
	data := LoginUserResponseToLoginDataResponse(profile, token)
	return response.Success(c, "Login berhasil", data)
}

func Register(c *fiber.Ctx) error {
	req := new(requests.RegisterRequest)
	if err := c.BodyParser(req); err != nil {
		return response.Error(c, 422, "Invalid request", nil)
	}

	if err := helpers.ValidateRegisterRequest(req); err != nil {
		return response.Error(c, 422, err.Error(), nil)
	}

	otp, user, contact, err := service.RegisterWithContact(req.FullName, req.Phone, req.Address, req.ContactSiteID, req.Email, req.Password, req.RoleID, req.SiteID)
	if err != nil {
		return response.Error(c, 500, err.Error(), nil)
	}

	resp := struct {
		User    interface{}       `json:"user"`
		Contact *contacts.Contact `json:"contact"`
		OTP     string            `json:"otp"`
	}{
		User:    UserToUserDTO(user),
		Contact: contact,
		OTP:     otp,
	}

	return response.Created(c, "User registered. Please verify using OTP sent.", resp)
}

func VerifyOTP(c *fiber.Ctx) error {
	req := new(requests.VerifyOTPRequest)
	if err := c.BodyParser(req); err != nil {
		return response.Error(c, 422, "Invalid request", nil)
	}

	if err := helpers.ValidateVerifyOTPRequest(req); err != nil {
		return response.Error(c, 422, err.Error(), nil)
	}

	err := service.VerifyOTP(req.Email, req.OTPCode)
	if err != nil {
		return response.Error(c, 400, err.Error(), nil)
	}

	return response.Success(c, "OTP verified successfully. Your account is now pending admin approval.", nil)
}

// LoginGoogle handles Google OAuth login or registration.
func LoginGoogle(c *fiber.Ctx) error {
	req := new(requests.GoogleLoginRequest)
	if err := c.BodyParser(req); err != nil {
		return response.Error(c, 422, "Invalid request", nil)
	}

	if err := helpers.ValidateGoogleLoginRequest(req); err != nil {
		return response.Error(c, 422, err.Error(), nil)
	}

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

	data := LoginUserResponseToLoginDataResponse(profile, token)
	return response.Success(c, "Google Login berhasil", data)
}

func SubmitGoogleUsername(c *fiber.Ctx) error {
	req := new(requests.SubmitGoogleUsernameRequest)
	if err := c.BodyParser(req); err != nil {
		return response.Error(c, 422, "Invalid request", nil)
	}

	if err := helpers.ValidateSubmitGoogleUsername(req); err != nil {
		return response.Error(c, 422, err.Error(), nil)
	}

	err := service.SubmitGoogleUsername(req.UserID, req.Username, req.FullName, req.Phone, req.Address, req.SiteID)
	if err != nil {
		return response.Error(c, 400, err.Error(), nil)
	}

	return response.Success(c, "Username updated and contact created. Your account is now pending admin approval.", nil)
}

// GetPendingApprovals list users waiting for admin approval
func GetPendingApprovals(c *fiber.Ctx) error {
	users, err := service.GetPendingApprovals()
	if err != nil {
		return response.Error(c, 500, err.Error(), nil)
	}

	dtos := UsersToUserDTOs(users)
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

// Logout logs out the user
func Logout(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(int64)
	if !ok {
		return response.Error(c, 400, "Invalid user ID", nil)
	}

	err := service.Logout(userID)
	if err != nil {
		return response.Error(c, 500, err.Error(), nil)
	}

	return response.Success(c, "Logout berhasil", nil)
}
