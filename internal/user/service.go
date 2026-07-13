package user

import "go/kir-tube/pkg/password"

type UserServiceDeps struct {
	UserRepository *UserRepository
}

type UserService struct {
	UserRepository *UserRepository
}

func NewUserService(deps *UserServiceDeps) *UserService {
	return &UserService{
		UserRepository: deps.UserRepository,
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

func (s *UserService) GetById(id string) (*User, error) {
	user, err := s.UserRepository.FindById(id)
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
