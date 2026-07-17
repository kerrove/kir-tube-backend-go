package user

import (
	"go/kir-tube/pkg/db"
)

type UserRepository struct {
	Database *db.Db
}

func NewUserRepository(database *db.Db) *UserRepository {
	return &UserRepository{Database: database}
}
func (repo *UserRepository) Create(user *User) (*User, error) {
	result := repo.Database.DB.Create(user)

	if result.Error != nil {
		return nil, result.Error
	}

	return user, nil
}
func (repo *UserRepository) Update(body *User) (*User, error) {

	result := repo.Database.DB.Save(body)

	user, err := repo.FindById(body.ID)
	if err != nil {
		return nil, err
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return user, nil
}
func (repo *UserRepository) FindByEmail(email string) (*User, error) {
	var user User
	res := repo.Database.DB.First(&user, "email = ?", email)

	if res.Error != nil {
		return nil, res.Error
	}

	return &user, nil
}
func (repo *UserRepository) FindById(id string) (*User, error) {
	var user User
	res := repo.Database.DB.First(&user, "id = ?", id)

	if res.Error != nil {
		return nil, res.Error
	}

	return &user, nil
}
func (repo *UserRepository) FindByVerifyToken(token string) (*User, error) {
	var user User
	res := repo.Database.DB.First(&user, "verification_token = ?", token)

	if res.Error != nil {
		return nil, res.Error
	}

	return &user, nil
}
func (repo *UserRepository) FindAll() []User {
	var users []User

	repo.Database.Table("users").
		Where("deleted_at is null").
		Select("name,email,id").
		Order("created_at desc").
		Scan(&users)

	return users
}
