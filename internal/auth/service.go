package auth

import (
	"errors"
	"go/kir-tube/configs"
	"go/kir-tube/internal/user"
	"go/kir-tube/pkg/jwt"
	"go/kir-tube/pkg/password"

	"net/http"
	"time"
)

type IUserService interface {
	Create(email, password string) (*user.User, error)
	GetById(string) (*user.User, error)
	GetByEmail(string) (*user.User, error)
}
type UserTokens struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type AuthServiceDeps struct {
	UserService IUserService
	Config      *configs.Config
}

type AuthService struct {
	UserService IUserService
	Config      *configs.Config
}

func NewAuthService(deps *AuthServiceDeps) *AuthService {
	return &AuthService{
		UserService: deps.UserService,
		Config:      deps.Config,
	}
}
func (s *AuthService) Login(body AuthRequest) (*AuthUser, error) {
	user, err := s.validateUser(body)

	if err != nil {
		return &AuthUser{}, err
	}

	resp, err := s.buildResponseObject(user)

	if err != nil {
		return &AuthUser{}, err
	}

	return resp, nil

}
func (s *AuthService) Register(body AuthRequest) (*AuthUser, error) {
	_, err := s.validateUser(body)
	if err == nil {
		return &AuthUser{}, errors.New(ErrUserExist)
	}

	user, err := s.UserService.Create(body.Email, body.Password)
	if err != nil {
		return &AuthUser{}, err
	}

	resp, err := s.buildResponseObject(user)
	if err != nil {
		return &AuthUser{}, err
	}

	return resp, nil
}

func (s *AuthService) validateUser(body AuthRequest) (*user.User, error) {
	us, err := s.UserService.GetByEmail(body.Email)

	if err != nil {
		return &user.User{}, err
	}

	isValid := password.Validate(us.Password, body.Password)

	if !isValid {
		return &user.User{}, errors.New(ErrWrongCredential)
	}

	return us, nil
}
func (s *AuthService) issueTokens(userId string) (*UserTokens, error) {
	payload := &jwt.JWTData{Id: userId, IsAdmin: false}

	accessToken, err := jwt.NewJWT(s.Config.Auth.Secret).Create(*payload, "1h")
	if err != nil {
		return &UserTokens{}, err
	}
	refreshToken, err := jwt.NewJWT(s.Config.Auth.Secret).Create(*payload, "720h")
	if err != nil {
		return &UserTokens{}, err
	}

	return &UserTokens{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}
func (s *AuthService) buildResponseObject(user *user.User) (*AuthUser, error) {
	tokens, err := s.issueTokens(user.ID)
	if err != nil {
		return &AuthUser{}, err
	}
	return &AuthUser{User: *user, AccessToken: tokens.AccessToken, RefreshToken: tokens.RefreshToken}, nil
}
func (s *AuthService) GetNewTokens(refreshToken string) (*AuthUser, error) {
	isValid, data := jwt.NewJWT(s.Config.Auth.Secret).Parse(refreshToken)
	if !isValid {
		return &AuthUser{}, errors.New("Unauthorized")
	}
	user, err := s.UserService.GetById(data.Id)

	if err != nil {
		return &AuthUser{}, err
	}

	res, err := s.buildResponseObject(user)

	if err != nil {
		return &AuthUser{}, err
	}

	return res, nil

}
func (s *AuthService) addRefreshTokenToResponse(w http.ResponseWriter, refreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenName,
		Value:    refreshToken,
		Path:     "/",
		Expires:  time.Now().AddDate(0, 0, ExpireDayRefreshToken),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *AuthService) RemoveRefreshTokenFromResponse(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}
