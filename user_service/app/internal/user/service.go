package user

import (
	"context"
	"errors"
	"fmt"
	"gitlab.konstweb.ru/ow/arch/notes/user_service/internal/apperror"
	"gitlab.konstweb.ru/ow/arch/notes/user_service/pkg/logging"
	"golang.org/x/crypto/bcrypt"
)

var _ Service = &service{}

type service struct {
	storage Storage
	logger  logging.Logger
}

func NewService(userStorage Storage, logger logging.Logger) (Service, error) {
	return &service{
		storage: userStorage,
		logger:  logger,
	}, nil
}

type Service interface {
	Create(ctx context.Context, dto CreateUserDTO) (string, error)
	GetByEmailAndPassword(ctx context.Context, email, password string) (User, error)
	GetOne(ctx context.Context, uuid string) (User, error)
	Update(ctx context.Context, dto UpdateUserDTO) error
	Delete(ctx context.Context, uuid string) error
}

func (s service) Create(ctx context.Context, dto CreateUserDTO) (userUUID string, err error) {
	if dto.Password != dto.RepeatPassword {
		return userUUID, apperror.BadRequestError("password does not match repeat password")
	}

	user := NewUser(dto)

	err = user.GeneratePasswordHash()
	if err != nil {
		s.logger.Errorf("failed to create user due to error %v", err)
		return
	}

	userUUID, err = s.storage.Create(ctx, user)

	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return userUUID, err
		}
		return userUUID, fmt.Errorf("failed to create user. error: %w", err)
	}

	return userUUID, nil
}

func (s service) GetByEmailAndPassword(ctx context.Context, email, password string) (u User, err error) {
	u, err = s.storage.FindByEmail(ctx, email)

	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return u, err
		}
		return u, fmt.Errorf("failed to find user by email. error: %w", err)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return u, apperror.ErrNotFound
	}

	return u, nil
}

func (s service) GetOne(ctx context.Context, uuid string) (u User, err error) {
	u, err = s.storage.FindOne(ctx, uuid)

	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return u, err
		}
		return u, fmt.Errorf("failed to find user by uuid. error: %w", err)
	}
	return u, nil
}

func (s service) Update(ctx context.Context, dto UpdateUserDTO) error {
	var updatedUser User
	if dto.OldPassword != dto.NewPassword || dto.NewPassword == "" {
		user, err := s.GetOne(ctx, dto.UUID)
		if err != nil {
			return err
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(dto.OldPassword))
		if err != nil {
			return apperror.BadRequestError("old password does not match current password")
		}

		dto.Password = dto.NewPassword
	}

	updatedUser = UpdatedUser(dto)

	err := updatedUser.GeneratePasswordHash()
	if err != nil {
		return fmt.Errorf("failed to update user. error %w", err)
	}

	err = s.storage.Update(ctx, updatedUser)

	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return err
		}
		return fmt.Errorf("failed to update user. error: %w", err)
	}
	return nil
}

func (s service) Delete(ctx context.Context, uuid string) error {
	err := s.storage.Delete(ctx, uuid)

	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return err
		}
		return fmt.Errorf("failed to delete user. error: %w", err)
	}
	return err
}
