package ports

import "go-hexagonal-api/internal/core/domain"

// UserRepository - Puerto Dirigido (Driven Port / SPI): Implementado por la infraestructura (DB)
type UserRepository interface {
	Create(user *domain.User) error
	GetByID(id int64) (*domain.User, error)
	GetByEmail(email string) (*domain.User, error)
	Update(user *domain.User) error
	Delete(id int64) error
}

// UserService - Puerto Conductor (Driving Port / API): Implementado por el núcleo del negocio
type UserService interface {
	Register(input domain.RegisterInput) (*domain.User, error)
	Login(input domain.LoginInput) (string, error)
	GetUserByID(id int64) (*domain.User, error)
	UpdateUser(id int64, input domain.UpdateInput) (*domain.User, error)
	DeleteUser(id int64) error
	UploadFile(file []byte, filename string) (string, error) // Método agregado para la subida de archivos
}