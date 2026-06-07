package domain

import (
	"errors"
	"time"
)

var (
	ErrUserNotFound      = errors.New("usuario no encontrado")
	ErrEmailAlreadyExists = errors.New("el correo electrónico ya está registrado")
	ErrInvalidCredentials = errors.New("credenciales inválidas")
)

type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Ocultar en las respuestas JSON
	CreatedAt time.Time `json:"created_at"`
}

type RegisterInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UpdateInput struct {
	Name  *string `json:"name"`
	Email *string `json:"email"`
}