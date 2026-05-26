package response

type Address struct {
	Province string `json:"province"`
	City     string `json:"city"`
	Detail   string `json:"detail"`
}

type Profile struct {
	Avatar string `json:"avatar"`
	Bio    string `json:"bio"`
}

type UserDetail struct {
	Username string  `json:"username"`
	Email    string  `json:"email"`
	Age      int     `json:"age"`
	Address  Address `json:"address"`
	Profile  Profile `json:"profile"`
}

type CompleteUserInfo struct {
	ID       string  `json:"id"`
	Username string  `json:"username"`
	Email    string  `json:"email"`
	Age      int     `json:"age"`
	Address  Address `json:"address"`
	Profile  Profile `json:"profile"`
}
