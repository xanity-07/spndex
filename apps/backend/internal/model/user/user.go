package user

import (
	"github.com/xanity-07/spndex/internal/enums"
	"github.com/xanity-07/spndex/internal/model"
)

type User struct {
	FirstName string          `json:"firstName" db:"first_name"`
	LastName  string          `json:"lastName" db:"last_name"`
	Email     string          `json:"email" db:"email"`
	Role      enums.UserRoles `json:"role" db:"user_role"`
	Password  string          `json:"-" db:"password_hash"`
	model.Base
}
