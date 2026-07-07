package services

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"supplierhub-backend/config"
	"supplierhub-backend/dto"
	"supplierhub-backend/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService handles operations related to users
type UserService struct {
	db *gorm.DB
}

// NewUserService creates a new UserService instance
func NewUserService() *UserService {
	return &UserService{db: config.DB}
}

// RegisterNewUser processes the registration of a new user (UMKM or Supplier)
func (s *UserService) RegisterNewUser(input dto.RegisterInput) (*models.User, error) {
	// 1. Cek apakah email sudah terdaftar
	var existingUser models.User
	if err := s.db.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("Email sudah terdaftar!")
	}

	// 2. File Upload Handling (Dokumen)
	var documentURL string
	if input.Document != nil {
		// Buat folder jika belum ada
		if err := os.MkdirAll(config.DocumentUploadDir, os.ModePerm); err != nil {
			return nil, errors.New("Gagal menyiapkan direktori penyimpanan")
		}

		// Gunakan timestamp untuk nama file yang unik
		filename := fmt.Sprintf("%d_%s", time.Now().Unix(), input.Document.Filename)
		filepath := config.DocumentUploadDir + "/" + filename

		src, err := input.Document.Open()
		if err != nil {
			return nil, errors.New("Gagal membuka dokumen legalitas")
		}
		defer src.Close()

		dst, err := os.Create(filepath)
		if err != nil {
			return nil, errors.New("Gagal menyimpan dokumen legalitas")
		}
		defer dst.Close()

		if _, err = io.Copy(dst, src); err != nil {
			return nil, errors.New("Gagal menulis file dokumen")
		}

		documentURL = filepath
	} else if input.Role == "supplier" {
		// Supplier wajib upload dokumen
		return nil, errors.New("Dokumen legalitas (SIUP/Akta) wajib diunggah untuk pendaftaran supplier")
	}

	// 3. Hash Password (atau dummy jika Google)
	var hashedPassword []byte
	var errHash error
	if input.IsGoogle {
		hashedPassword, errHash = bcrypt.GenerateFromPassword([]byte("google-oauth-dummy-pass"), bcrypt.DefaultCost)
	} else {
		if input.Password == "" {
			return nil, errors.New("Password wajib diisi!")
		}
		hashedPassword, errHash = bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	}

	if errHash != nil {
		return nil, errors.New("Gagal memproses password")
	}

	// 4. Simpan ke Database
	newUser := models.User{
		BusinessName: input.BusinessName,
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
		Role:         models.Role(input.Role),
		Address:      input.Address,
		Category:     input.Category,
		Region:       input.Region,
		DocumentURL:  documentURL,
		Status:       "pending", // Default status, membutuhkan verifikasi admin
	}

	if err := s.db.Create(&newUser).Error; err != nil {
		return nil, errors.New("Gagal menyimpan data pengguna")
	}

	return &newUser, nil
}
