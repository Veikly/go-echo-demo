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
