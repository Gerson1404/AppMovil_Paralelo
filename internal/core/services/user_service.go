package services

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go-hexagonal-api/internal/core/domain"
	"go-hexagonal-api/internal/core/ports"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	repo      ports.UserRepository
	jwtSecret []byte
}

func NewUserService(repo ports.UserRepository, secret string) ports.UserService {
	return &userService{
		repo:      repo,
		jwtSecret: []byte(secret),
	}
}

func (s *userService) Register(input domain.RegisterInput) (*domain.User, error) {
	existing, _ := s.repo.GetByEmail(input.Email)
	if existing != nil {
		return nil, domain.ErrEmailAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Name:      input.Name,
		Email:     input.Email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) Login(input domain.LoginInput) (string, error) {
	user, err := s.repo.GetByEmail(input.Email)
	if err != nil {
		return "", domain.ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		return "", domain.ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(), 
	})

	return token.SignedString(s.jwtSecret)
}

func (s *userService) GetUserByID(id int64) (*domain.User, error) {
	return s.repo.GetByID(id)
}

func (s *userService) UpdateUser(id int64, input domain.UpdateInput) (*domain.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.Email != nil {
		if *input.Email != user.Email {
			existing, _ := s.repo.GetByEmail(*input.Email)
			if existing != nil {
				return nil, domain.ErrEmailAlreadyExists
			}
			user.Email = *input.Email
		}
	}

	if err := s.repo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) DeleteUser(id int64) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	return s.repo.Delete(id)
}

func (s *userService) UploadFile(fileBytes []byte, filename string) (string, error) {
	// MODIFICACIÓN 1: Usar la carpeta temporal permitida por AWS
	uploadDir := "/tmp"

	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("no se pudo crear el directorio de subida: %w", err)
	}

	ext := filepath.Ext(filename)
	baseName := filename[:len(filename)-len(ext)]
	uniqueFilename := fmt.Sprintf("%s_%d%s", baseName, time.Now().Unix(), ext)
	
	filePath := filepath.Join(uploadDir, uniqueFilename)

	if err := os.WriteFile(filePath, fileBytes, 0644); err != nil {
		return "", fmt.Errorf("no se pudo guardar el archivo: %w", err)
	}

	// MODIFICACIÓN 2: Usar la URL pública real para que Android la pueda renderizar
	baseURL := "https://mvv94io5s5.execute-api.us-east-1.amazonaws.com"
	fileURL := fmt.Sprintf("%s/uploads/%s", baseURL, uniqueFilename)
	
	return fileURL, nil
}