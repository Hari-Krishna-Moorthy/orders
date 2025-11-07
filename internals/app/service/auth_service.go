package service

import (
	"sync"

	"github.com/Hari-Krishna-Moorthy/orders/internals/app/helpers"
	"github.com/Hari-Krishna-Moorthy/orders/internals/app/models"
	"github.com/Hari-Krishna-Moorthy/orders/internals/app/repository"
	types "github.com/Hari-Krishna-Moorthy/orders/internals/app/types"
	"github.com/Hari-Krishna-Moorthy/orders/internals/platform/config"
)

type AuthService struct{ repo *repository.UserRepository }

var (
	authSvc *AuthService
	once    sync.Once
)

func NewAuthService(repo *repository.UserRepository) *AuthService {
	once.Do(func() { authSvc = &AuthService{repo: repo} })
	return authSvc
}

func GetAuthService() *AuthService { return authSvc }

type AuthResponse struct {
	AccessToken string `json:"access_token"`
}

func (s *AuthService) SignUp(in types.SignUpInput) (*AuthResponse, error) {
	if in.Email == "" || in.Password == "" {
		return nil, types.EMAIL_AND_PASSWORD_REQUIRED_ERROR
	}
	if _, err := s.repo.FindByEmail(in.Email); err == nil {
		return nil, types.EMAIL_ALREADY_REGISTERED_ERROR
	}
	hash, err := helpers.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}
	u := &models.User{Email: in.Email, Name: in.Name, Hash: hash}
	if err := s.repo.Create(u); err != nil {
		return nil, err
	}
	cfg := config.Get()
	t, err := helpers.CreateAccessToken(cfg.JWT.AccessSecret, cfg.JWT.Issuer, cfg.JWT.AccessTTLMinutes, u.ID)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{AccessToken: t}, nil
}

func (s *AuthService) SignIn(in types.SignInInput) (*AuthResponse, error) {
	u, err := s.repo.FindByEmail(in.Email)
	if err != nil {
		return nil, types.INVALID_CREDENTIALS_ERROR
	}
	if !helpers.CheckPassword(u.Hash, in.Password) {
		return nil, types.INVALID_CREDENTIALS_ERROR
	}
	cfg := config.Get()
	t, err := helpers.CreateAccessToken(cfg.JWT.AccessSecret, cfg.JWT.Issuer, cfg.JWT.AccessTTLMinutes, u.ID)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{AccessToken: t}, nil
}
