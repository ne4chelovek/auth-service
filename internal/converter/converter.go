package converter

import (
	"database/sql"
	"github.com/ne4chelovek/auth-service/internal/model"
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
		ID:       user.GetId(),
		Name:     convertStringValueToNullString(user.GetName()),
		Email:    convertStringValueToNullString(user.GetEmail()),
		Password: convertStringValueToNullString(user.GetPassword()),
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
