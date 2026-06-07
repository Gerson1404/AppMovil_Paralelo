package repositories

import (
	"database/sql"
	"errors"

	"go-hexagonal-api/internal/core/domain"
	"go-hexagonal-api/internal/core/ports"

	_ "github.com/lib/pq"
)

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) ports.UserRepository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(user *domain.User) error {
	query := `INSERT INTO users (name, email, password, created_at) 
              VALUES ($1, $2, $3, $4) RETURNING id`
	err := r.db.QueryRow(query, user.Name, user.Email, user.Password, user.CreatedAt).Scan(&user.ID)
	return err
}

func (r *postgresRepository) GetByID(id int64) (*domain.User, error) {
	query := `SELECT id, name, email, password, created_at FROM users WHERE id = $1`
	row := r.db.QueryRow(query, id)

	var user domain.User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *postgresRepository) GetByEmail(email string) (*domain.User, error) {
	query := `SELECT id, name, email, password, created_at FROM users WHERE email = $1`
	row := r.db.QueryRow(query, email)

	var user domain.User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *postgresRepository) Update(user *domain.User) error {
	query := `UPDATE users SET name = $1, email = $2 WHERE id = $3`
	_, err := r.db.Exec(query, user.Name, user.Email, user.ID)
	return err
}

func (r *postgresRepository) Delete(id int64) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}