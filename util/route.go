package util

const (
	CreateUser = "/register"
	LoginUser  = "/login"
	GetUser    = "/user"

	CreateAccount = "/account"
	GetAccount    = "/account/:id"
	ListAccount   = "/accounts"

	CreateEntry = "/entries"
	GetEntry    = "/entries/:id"
	ListEntry   = "/entries"

	CreateTransfer = "/transfers"
	GetTransfer    = "/transfer/:id"
	ListTransfer   = "/transfers"

	RenewToken = "/tokens/renew_access"
)
