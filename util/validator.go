package util

const (
	USD  = "USD"
	EURO = "EURO"
	SGD  = "SGD"

	ADMIN = "Admin"
	USER  = "User"
)

func IsSupportedCurrency(currency string) bool {

	switch currency {
	case USD, EURO, SGD:
		return true
	}
	return false
}

func IsSupportedRoles(role string) bool {
	switch role {
	case ADMIN, USER:
		return true

	}
	return false
}
