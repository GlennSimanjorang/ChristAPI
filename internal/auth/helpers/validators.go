package helpers

import (
	"christ-api/internal/auth/dto/requests"
	"errors"
)

// ValidateLoginRequest validates login request
func ValidateLoginRequest(req *requests.LoginRequest) error {
	if req.Email == "" {
		return errors.New("email is required")
	}
	if req.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

// ValidateRegisterRequest validates register request
func ValidateRegisterRequest(req *requests.RegisterRequest) error {
	errs := make(map[string][]string)

	if req.FullName == "" {
		errs["full_name"] = append(errs["full_name"], "Full name is required")
	}
	if req.Email == "" {
		errs["email"] = append(errs["email"], "Email is required")
	}
	if req.Password == "" {
		errs["password"] = append(errs["password"], "Password is required")
	}

	if len(errs) > 0 {
		return errors.New("validation failed")
	}

	return nil
}

// ValidateVerifyOTPRequest validates OTP verification request
func ValidateVerifyOTPRequest(req *requests.VerifyOTPRequest) error {
	if req.Email == "" {
		return errors.New("email is required")
	}
	if req.OTPCode == "" {
		return errors.New("otp_code is required")
	}
	return nil
}

// ValidateGoogleLoginRequest validates Google login request
func ValidateGoogleLoginRequest(req *requests.GoogleLoginRequest) error {
	if req.IDToken == "" {
		return errors.New("id_token is required")
	}
	return nil
}

// ValidateSubmitGoogleUsername validates submit username request
func ValidateSubmitGoogleUsername(req *requests.SubmitGoogleUsernameRequest) error {
	if req.UserID == 0 {
		return errors.New("user_id is required")
	}
	if req.Username == "" {
		return errors.New("username is required")
	}
	return nil
}
