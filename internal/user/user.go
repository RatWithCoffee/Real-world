package user

import (
	"log/slog"
	"realworld/internal/utils"
)

type User struct {
	Id                uint                    `json:"-"`
	Email             string                  `json:"email"`
	CreatedAt         string                  `json:"createdAt"`
	UpdatedAt         string                  `json:"updatedAt"`
	Username          string                  `json:"username"`
	Token             string                  `json:"token"`
	Password          string                  `json:"password"`
	EncryptedPassword []byte                  `json:"-"`
	Salt              []byte                  `json:"-"`
	Bio               utils.NullStringWrapper `json:"bio"`
}

func (user *User) GetId() uint {
	return user.Id
}

func (user *User) SetToken(token string) {
	user.Token = token
}

func (user User) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Group("user",
			slog.Uint64("id", uint64(user.Id)),
			slog.String("name", user.Username),
			slog.String("email", user.Email),
			slog.String("createdAt", user.CreatedAt),
			slog.String("updatedAt", user.UpdatedAt),
			slog.String("bio", user.Bio.String)))
}
