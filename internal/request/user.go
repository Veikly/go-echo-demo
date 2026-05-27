package request

import "go-echo-demo/internal/usecase/usecaseio"

type Address struct {
	Province string `json:"province"`
	City     string `json:"city"`
	Detail   string `json:"detail"`
}

type Profile struct {
	Avatar string `json:"avatar"`
	Bio    string `json:"bio"`
}

type CompleteUserInfoInput struct {
	Username string  `json:"username"`
	Age      int     `json:"age"`
	Address  Address `json:"address"`
	Profile  Profile `json:"profile"`
}

func (input *CompleteUserInfoInput) ToCompleteUserInfoInput() usecaseio.CompleteUserInfoDetail {
	return usecaseio.CompleteUserInfoDetail{
		Username: input.Username,
		Age:      input.Age,
		Address:  usecaseio.Address(input.Address),
		Profile:  usecaseio.Profile(input.Profile),
	}
}
