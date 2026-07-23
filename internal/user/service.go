package user

import (
	"go/kir-tube/pkg/di"
	"go/kir-tube/pkg/password"
)

type UserProfile struct {
	User
	SubscribedVideos []di.SubscribedVideo     `json:"subscribedVideos"`
	Likes            []di.SubscribedVideo     `json:"likes"`
	Subscriptions    []di.SubscriptionChannel `json:"subscriptions"`
}

// UserRepositoryPort is the persistence contract the user service depends on.
// Depending on the interface (not *UserRepository) keeps the service testable
// with a mock. *UserRepository satisfies it.
type UserRepositoryPort interface {
	Create(user *User) (*User, error)
	Update(body *User) (*User, error)
	FindById(id string) (*User, error)
	FindByEmail(email string) (*User, error)
	FindByVerifyToken(token string) (*User, error)
	FindAll() []User
}

type UserServiceDeps struct {
	UserRepository  UserRepositoryPort
	VideoRepository di.IVideoRepository
}

type UserService struct {
	UserRepository  UserRepositoryPort
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

// UpdateProfile applies a partial change to the user's own record and returns
// the refreshed profile. Password is re-hashed when provided.
func (s *UserService) UpdateProfile(userID string, req *UpdateProfileReq) (*UserProfile, error) {
	user, err := s.GetById(userID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		user.Name = req.Name
	}
	if req.Password != nil && *req.Password != "" {
		hashed, err := password.Encode(*req.Password)
		if err != nil {
			return nil, err
		}
		user.Password = hashed
	}

	if _, err := s.UserRepository.Update(user); err != nil {
		return nil, err
	}

	return s.GetProfile(userID)
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

	likedVideos, err := s.VideoRepository.FindLikedVideos(userID)
	if err != nil {
		return nil, err
	}

	subscriptions, err := s.VideoRepository.FindSubscriptions(userID)
	if err != nil {
		return nil, err
	}

	return &UserProfile{
		User:             *user,
		SubscribedVideos: subscribedVideos,
		Likes:            likedVideos,
		Subscriptions:    subscriptions,
	}, nil
}
