package user

import (
	"go/kir-tube/pkg/di"
	"go/kir-tube/pkg/password"
)

type UserProfile struct {
	User
	SubscribedVideos []di.SubscribedVideo `json:"subscribedVideos"`
}

type UserServiceDeps struct {
	UserRepository  *UserRepository
	VideoRepository di.IVideoRepository
}

type UserService struct {
	UserRepository  *UserRepository
	VideoRepository di.IVideoRepository
}

func NewUserService(deps *UserServiceDeps) *UserService {
	return &UserService{
		UserRepository:  deps.UserRepository,
		VideoRepository: deps.VideoRepository,
	}
}

func (s *UserService) Create(email, bodyPassword string) (*User, error) {
	hashedPassword, err := password.Encode(bodyPassword)

	if err != nil {
		return &User{}, err
	}

	user, err := s.UserRepository.Create(&User{
		Password: hashedPassword,
		Email:    email,
	})

	if err != nil {
		return &User{}, err
	}

	return user, nil
}

func (s *UserService) Update(body *User) (*User, error) {
	user, err := s.UserRepository.Update(body)

	if err != nil {
		return &User{}, err
	}

	return user, nil
}

func (s *UserService) GetById(id string) (*User, error) {
	user, err := s.UserRepository.FindById(id)
	if err != nil {
		return &User{}, err
	}
	return user, nil
}
func (s *UserService) GetByVerifyToken(token string) (*User, error) {
	user, err := s.UserRepository.FindByVerifyToken(token)
	if err != nil {
		return &User{}, err
	}

	return user, nil
}

func (s *UserService) GetByEmail(email string) (*User, error) {
	user, err := s.UserRepository.FindByEmail(email)
	if err != nil {
		return &User{}, err
	}
	return user, nil
}
func (s *UserService) GetAll() []User {
	users := s.UserRepository.FindAll()

	return users
}

func (s *UserService) GetProfile(userID string) (*UserProfile, error) {
	user, err := s.GetById(userID)
	if err != nil {
		return nil, err
	}
	subscribedVideos, err := s.VideoRepository.FindSubscribedVideos(userID)
	if err != nil {
		return nil, err
	}

	return &UserProfile{User: *user, SubscribedVideos: subscribedVideos}, nil
}
