package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/nwasiq/fieldops/backend/internal/models"
)

// MinPasswordLength is the shortest password accepted on create or update.
const MinPasswordLength = 8

// UserService manages users. It is the only code that reads or writes
// users.licence_number, and it always goes through the Cipher (§1.3).
type UserService struct {
	db     *gorm.DB
	cipher *Cipher
	auth   *AuthService
}

// NewUserService builds a UserService.
func NewUserService(db *gorm.DB, cipher *Cipher, auth *AuthService) *UserService {
	return &UserService{db: db, cipher: cipher, auth: auth}
}

// UserFilter narrows a user list.
type UserFilter struct {
	Role   string
	Active *bool
}

// UserInput is the writable shape of a user. Pointer fields are optional on
// update: nil leaves the stored value alone.
type UserInput struct {
	Email         string
	FirstName     string
	LastName      string
	Role          string
	IsActive      *bool
	LicenceNumber *string
	Password      *string
}

func (in *UserInput) validate() error {
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	in.FirstName = strings.TrimSpace(in.FirstName)
	in.LastName = strings.TrimSpace(in.LastName)
	switch {
	case in.Email == "" || !strings.Contains(in.Email, "@"):
		return invalidf("email is required")
	case in.FirstName == "":
		return invalidf("first_name is required")
	case in.LastName == "":
		return invalidf("last_name is required")
	case !models.IsValidRole(in.Role):
		return invalidf("role must be one of %s", strings.Join(models.AllRoles, ", "))
	case in.Password != nil && len(*in.Password) < MinPasswordLength:
		return invalidf("password must be at least %d characters", MinPasswordLength)
	}
	return nil
}

// List returns users that are not soft-deleted, filtered and paged (§1.2).
// Licence numbers are never included in a list.
func (s *UserService) List(ctx context.Context, filter UserFilter, page Page) (PageResult[models.User], error) {
	query := s.db.WithContext(ctx).Model(&models.User{})
	if filter.Role != "" {
		query = query.Where("role = ?", filter.Role)
	}
	if filter.Active != nil {
		query = query.Where("is_active = ?", *filter.Active)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return PageResult[models.User]{}, fmt.Errorf("count users: %w", err)
	}
	var users []models.User
	err := query.Order("last_name, first_name, id").Offset(page.Offset()).Limit(page.Size).Find(&users).Error
	if err != nil {
		return PageResult[models.User]{}, fmt.Errorf("list users: %w", err)
	}
	for i := range users {
		users[i].LicenceNumber = ""
	}
	return newPageResult(users, total, page), nil
}

// Get returns one user. The licence number is decrypted only when the caller
// is entitled to it; otherwise it is blanked so ciphertext never leaves the
// service (§1.3).
func (s *UserService) Get(ctx context.Context, id uint, withLicence bool) (*models.User, error) {
	var user models.User
	err := s.db.WithContext(ctx).First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, notFound("user")
	}
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if !withLicence {
		user.LicenceNumber = ""
		return &user, nil
	}
	if err := decryptUserSensitive(s.cipher, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// Create adds a user. A password is required on create.
func (s *UserService) Create(ctx context.Context, in UserInput) (*models.User, error) {
	if in.Password == nil {
		return nil, invalidf("password is required")
	}
	if err := in.validate(); err != nil {
		return nil, err
	}
	hash, err := s.auth.HashPassword(*in.Password)
	if err != nil {
		return nil, err
	}
	user := models.User{
		Email:        in.Email,
		PasswordHash: hash,
		FirstName:    in.FirstName,
		LastName:     in.LastName,
		Role:         in.Role,
		IsActive:     true,
	}
	if in.IsActive != nil {
		user.IsActive = *in.IsActive
	}
	if in.LicenceNumber != nil {
		user.LicenceNumber = strings.TrimSpace(*in.LicenceNumber)
	}
	if err := encryptUserSensitive(s.cipher, &user); err != nil {
		return nil, err
	}
	taken, err := s.emailTaken(ctx, user.Email, 0)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, conflict("email is already in use")
	}
	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	user.LicenceNumber = ""
	return &user, nil
}

// Update replaces a user's profile fields and, when supplied, their licence
// number, password and active flag.
func (s *UserService) Update(ctx context.Context, id uint, in UserInput) (*models.User, error) {
	if err := in.validate(); err != nil {
		return nil, err
	}
	var user models.User
	err := s.db.WithContext(ctx).First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, notFound("user")
	}
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	taken, err := s.emailTaken(ctx, in.Email, user.ID)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, conflict("email is already in use")
	}
	user.Email = in.Email
	user.FirstName = in.FirstName
	user.LastName = in.LastName
	user.Role = in.Role
	if in.IsActive != nil {
		user.IsActive = *in.IsActive
	}
	if in.LicenceNumber != nil {
		user.LicenceNumber = strings.TrimSpace(*in.LicenceNumber)
		if err := encryptUserSensitive(s.cipher, &user); err != nil {
			return nil, err
		}
	}
	if in.Password != nil {
		hash, err := s.auth.HashPassword(*in.Password)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = hash
	}
	if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	user.LicenceNumber = ""
	return &user, nil
}

// Delete soft-deletes a user. Their attributions stay readable through
// unscoped preloads (§1.2). An actor cannot delete their own account.
func (s *UserService) Delete(ctx context.Context, id uint, actor Actor) error {
	if id == actor.ID {
		return conflict("you cannot delete your own account")
	}
	result := s.db.WithContext(ctx).Delete(&models.User{}, id)
	if result.Error != nil {
		return fmt.Errorf("delete user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return notFound("user")
	}
	return nil
}

// emailTaken checks the unique email across live and soft-deleted rows, since
// the database index covers both.
func (s *UserService) emailTaken(ctx context.Context, email string, exceptID uint) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).Unscoped().Model(&models.User{}).
		Where("email = ? AND id <> ?", email, exceptID).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("check email: %w", err)
	}
	return count > 0, nil
}
