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
	// Verificar si el email ya existe
	existing, _ := s.repo.GetByEmail(input.Email)
	if existing != nil {
		return nil, domain.ErrEmailAlreadyExists
	}

	// Hashear la contraseña
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

	// Validar la contraseña
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		return "", domain.ErrInvalidCredentials
	}

	// Generar Token JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(), // Expira en 24h
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
		// Validar que el nuevo email no esté duplicado
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

// Implementación del nuevo método UploadFile
func (s *userService) UploadFile(fileBytes []byte, filename string) (string, error) {
	uploadDir := "uploads"

	// 1. Asegurar que el directorio de uploads exista (Crea la carpeta si no está)
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("no se pudo crear el directorio de subida: %w", err)
	}

	// 2. Generar un nombre de archivo único agregando un timestamp
	// Esto evita que si se suben dos archivos llamados "foto.png" se sobreescriban
	ext := filepath.Ext(filename)
	baseName := filename[:len(filename)-len(ext)]
	uniqueFilename := fmt.Sprintf("%s_%d%s", baseName, time.Now().Unix(), ext)
	
	// 3. Crear la ruta final del archivo
	filePath := filepath.Join(uploadDir, uniqueFilename)

	// 4. Escribir los bytes en el archivo físico
	if err := os.WriteFile(filePath, fileBytes, 0644); err != nil {
		return "", fmt.Errorf("no se pudo guardar el archivo: %w", err)
	}

	// 5. Retornar la URL donde será accesible. 
	// Nota: Si vas a consumirlo desde el emulador de Android (10.0.2.2), puedes ajustar esta URL 
	// dinámicamente mediante variables de entorno si lo prefieres.
	fileURL := fmt.Sprintf("http://localhost:8080/uploads/%s", uniqueFilename)
	
	return fileURL, nil
}