package converter

import (
	"database/sql"
	"github.com/ne4chelovek/auth-service/internal/model"
	descAuth "github.com/ne4chelovek/auth-service/pkg/auth_v1"
	descUsers "github.com/ne4chelovek/auth-service/pkg/users_v1"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

func FromDescCreateToUser(user *descUsers.CreateUser) *model.CreateUser {
	return &model.CreateUser{
		Name:            user.Name,
		Email:           user.Email,
		Password:        user.Password,
		PasswordConfirm: user.PasswordConfirm,
		Role:            descUsers.Role_name[int32(user.Role)],
	}
}

func FromUserToDesc(user *model.User) *descUsers.User {
	return &descUsers.User{
		Id:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      descUsers.Role(descUsers.Role_value[user.Role]),
		CreatedAt: user.CreatedAt,
	}
}

func FromDescUpdateToAuth(user *descUsers.UpdateUser) *model.UpdateUser {
	return &model.UpdateUser{
		ID:       user.GetId(),                                       // int64 -> int64, конвертация не требуется
		Name:     convertStringValueToNullString(user.GetName()),     // *wrapperspb.StringValue -> sql.NullString
		Email:    convertStringValueToNullString(user.GetEmail()),    // *wrapperspb.StringValue -> sql.NullString
		Password: convertStringValueToNullString(user.GetPassword()), // *wrapperspb.StringValue -> sql.NullString
	}
}

func FromAuthDescToLogin(req *descAuth.Login) *model.UserCreds {
	return &model.UserCreds{
		UserNames: req.Usernames,
		Password:  req.Password,
	}
}

// Вспомогательные функции для конвертации

// convertStringValueToNullString конвертирует *wrapperspb.StringValue в sql.NullString
func convertStringValueToNullString(sv *wrapperspb.StringValue) sql.NullString {
	if sv == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{
		String: sv.GetValue(),
		Valid:  true,
	}
}
