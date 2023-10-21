package store

import (
	"context"
)

// Role is the type of a role.
type Role string

const (
	// RoleHost is the HOST role.
	RoleHost Role = "HOST"
	// RoleAdmin is the ADMIN role.
	RoleAdmin Role = "ADMIN"
	// RoleUser is the USER role.
	RoleUser Role = "USER"

	RoleExternal Role = "EXTERNAL"
)

func (e Role) String() string {
	switch e {
	case RoleHost:
		return "HOST"
	case RoleAdmin:
		return "ADMIN"
	case RoleUser:
		return "USER"
	case RoleExternal:
		return "EXTERNAL"
	}
	return "USER"
}

type User struct {
	ID int32

	// Standard fields
	RowStatus RowStatus
	CreatedTs int64
	UpdatedTs int64

	// Domain specific fields
	Username     string
	Role         Role
	Email        string
	Nickname     string
	PasswordHash string
	AvatarURL    string
}

type UpdateUser struct {
	ID int32

	UpdatedTs    *int64
	RowStatus    *RowStatus
	Username     *string
	Role         *Role
	Email        *string
	Nickname     *string
	Password     *string
	AvatarURL    *string
	PasswordHash *string
}

type FindUser struct {
	ID        *int32
	RowStatus *RowStatus
	Username  *string
	Role      *Role
	Email     *string
	Nickname  *string
}

type DeleteUser struct {
	ID int32
}

func (s *Store) CreateUser(ctx context.Context, create *User) (*User, error) {
	user, err := s.driver.CreateUser(ctx, create)
	if err != nil {
		return nil, err
	}

	s.userCache.Store(user.ID, user)
	return user, nil
}

func (s *Store) UpdateUser(ctx context.Context, update *UpdateUser) (*User, error) {
	user, err := s.driver.UpdateUser(ctx, update)
	if err != nil {
		return nil, err
	}

	s.userCache.Store(user.ID, user)
	return user, nil
}

func (s *Store) ListUsers(ctx context.Context, find *FindUser) ([]*User, error) {
	list, err := s.driver.ListUsers(ctx, find)
	if err != nil {
		return nil, err
	}

	for _, user := range list {
		s.userCache.Store(user.ID, user)
	}
	return list, nil
}

func (s *Store) GetUser(ctx context.Context, find *FindUser) (*User, error) {
	if find.ID != nil {
		if cache, ok := s.userCache.Load(*find.ID); ok {
			return cache.(*User), nil
		}
	}

	list, err := s.ListUsers(ctx, find)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil
	}

	user := list[0]
	s.userCache.Store(user.ID, user)
	return user, nil
}

func (s *Store) DeleteUser(ctx context.Context, delete *DeleteUser) error {
	err := s.driver.DeleteUser(ctx, delete)
	if err != nil {
		return err
	}

	s.userCache.Delete(delete.ID)
	return nil
}
