package util

const (
	CreateUser = "/register"
	LoginUser  = "/login"
	GetUser    = "/user"

	CreateAccount = "/account"
	GetAccount    = "/account/:id"
	ListAccount   = "/accounts"

	GetEntry  = "/entries/:id"
	ListEntry = "/entries"

	CreateTransfer = "/transfers"
	GetTransfer    = "/transfer/:id"
	ListTransfer   = "/transfers"

	RenewToken  = "/tokens/renew_access"
	VerifyEmail = "/verify_email"
)
