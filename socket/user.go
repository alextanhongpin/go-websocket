package socket

// UserID represents the unique user identifier for authenticated users.
type UserID string

func (u UserID) String() string {
	return string(u)
}

// NewUserID returns a new UserID type.
func NewUserID(id string) UserID {
	return UserID(id)
}
