package model

import (
	"database/sql"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type User struct {
	ID        int64
	Name      string
	Email     string
	Role      string
	CreatedAt *timestamppb.Timestamp
}

type CreateUser struct {
	Name            string
	Email           string
	Password        string
	PasswordConfirm string
	Role            string
}

type UpdateUser struct {
	ID        int64
	Name      sql.NullString
	Email     sql.NullString
	Password  sql.NullString
	UpdatedAt sql.NullTime
}

type UserCreds struct {
	UserNames string `json:"userNames"`
	Password  string `json:"password"`
}

type UserInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}
