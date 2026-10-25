// Package infra connects the User context to Postgres.
package infra

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/rudradcruze/go-react-template/backend/internal/user/app"
	"github.com/rudradcruze/go-react-template/backend/internal/user/domain"
)

var _ app.UserRepository = (*UserRepository)(nil)

// userModel is the row of the table users. The domain does not see it.
type userModel struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	EmailAddress string
	DisplayName  string
	CreatedAt    time.Time
}

func (userModel) TableName() string { return "users" }

// toDomain builds the User of a row. The database has only values that the domain accepted, so it does not check them again.
func (m userModel) toDomain() domain.User {
	return domain.User{
		ID:           m.ID,
		EmailAddress: domain.EmailAddress(m.EmailAddress),
		DisplayName:  domain.DisplayName(m.DisplayName),
		CreatedAt:    m.CreatedAt.UTC(),
	}
}

// UserRepository stores Users in Postgres with GORM.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository builds a repository on a database connection.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create saves a new User. It returns domain.ErrEmailTaken when another User has the same email address.
func (r *UserRepository) Create(ctx context.Context, user domain.User) error {
	model := userModel{
		ID:           user.ID,
		EmailAddress: string(user.EmailAddress),
		DisplayName:  string(user.DisplayName),
		CreatedAt:    user.CreatedAt,
	}
	err := r.db.WithContext(ctx).Create(&model).Error
	// The email address is the only unique value that a caller chooses. The ID is a new UUID, so a clash of the ID does not occur.
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return domain.ErrEmailTaken
	}
	return err
}

// FindByID gives the User with this ID. It returns domain.ErrUserNotFound when no User has the ID.
func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	var model userModel
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.User{}, domain.ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, err
	}
	return model.toDomain(), nil
}
