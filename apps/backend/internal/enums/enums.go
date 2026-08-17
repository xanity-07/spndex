package enums

type ExpenseCategory string

const (
	FOOD          ExpenseCategory = "food"
	TRANSPORT     ExpenseCategory = "transport"
	UTILITIES     ExpenseCategory = "utilities"
	ENTERTAINMENT ExpenseCategory = "entertainment"
	HEALTHCARE    ExpenseCategory = "healthcare"
	SHOPPING      ExpenseCategory = "shopping"
	EDUCATION     ExpenseCategory = "education"
	OTHER         ExpenseCategory = "other"
)

// Valid returns true if its a valid ExpenseCategory
func (e ExpenseCategory) Valid() bool {
	switch e {
	case FOOD, TRANSPORT, UTILITIES, ENTERTAINMENT, HEALTHCARE, SHOPPING, EDUCATION, OTHER:
		return true
	}
	return false
}

type BindingSource string

const (
	BindingJSON  BindingSource = "json"
	BindingQuery BindingSource = "query"
	BindingURI   BindingSource = "uri"
)

type UserRoles string

const (
	UserRoleUser  UserRoles = "user"
	UserRoleAdmin UserRoles = "admin"
)

type RedisKeyPrefix string

const (
	SessionKeyPrefix RedisKeyPrefix = "session:"
)

func (p RedisKeyPrefix) Key(id string) string {
	return string(p) + id
}
