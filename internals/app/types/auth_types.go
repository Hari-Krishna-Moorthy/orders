package types

type SignUpInput struct {
	Email,
	Password,
	Name string
}

type SignInInput struct {
	Email,
	Password string
}

type SignUpRequest struct {
	Email,
	Password,
	Name string
}

type SignInRequest struct {
	Email,
	Password string
}

type GenericOK struct {
	Status string `json:"status"`
}
