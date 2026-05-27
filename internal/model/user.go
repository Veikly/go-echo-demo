package model

type Address struct {
	Province string
	City     string
	Detail   string
}

type Profile struct {
	Avatar string
	Bio    string
}

type User struct {
	ID       string
	Username string
	Email    string
	Age      int
	Address  Address
	Profile  Profile
}

func (u *User) ToMap() map[string]any {
	return map[string]any{
		"Username": u.Username,
		"Email":    u.Email,
		"Age":      u.Age,
		"Address":  u.Address,
		"Profile":  u.Profile,
	}
}
