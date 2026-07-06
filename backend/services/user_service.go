package services

import (
	"errors"
	"supplierhub-backend/config"
	"supplierhub-backend/models"
	"supplierhub-backend/repositories"

	"golang.org/x/crypto/bcrypt"
)

type RegisterDTO struct {
	BusinessName string
	Email        string
	Password     string
	Role         string
	Address      string
	Category     string
	Region       string
	DocumentURL  string
	IsGoogle     bool
}

type UserService struct {
	userRepo *repositories.UserRepository
}

func NewUserService() *UserService {
	return &UserService{
		userRepo: repositories.NewUserRepository(config.DB),
	}
}

func (s *UserService) RegisterNewUser(dto RegisterDTO) (*models.User, error) {
	// Cek email
	_, err := s.userRepo.FindByEmail(dto.Email)
	if err == nil {
		return nil, errors.New("Email sudah terdaftar!")
	}

	// Hashing password
	var hashedPassword []byte
	var errHash error
	if dto.IsGoogle {
		hashedPassword, errHash = bcrypt.GenerateFromPassword([]byte("google-oauth-dummy-pass"), bcrypt.DefaultCost)
	} else {
		hashedPassword, errHash = bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	}

	if errHash != nil {
		return nil, errHash
	}

	newUser := models.User{
		BusinessName: dto.BusinessName,
		Email:        dto.Email,
		PasswordHash: string(hashedPassword),
		Role:         models.Role(dto.Role),
		Address:      dto.Address,
		Category:     dto.Category,
		Region:       dto.Region,
		DocumentURL:  dto.DocumentURL,
		Status:       "pending",
	}

	err = s.userRepo.Create(&newUser)
	if err != nil {
		return nil, err
	}

	return &newUser, nil
}
