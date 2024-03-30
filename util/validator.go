package util

const (
	USD  = "USD"
	EURO = "EURO"
	SGD  = "SGD"

	DEPOSITOR = "depositor"
	BANKER    = "banker"
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
	case DEPOSITOR, BANKER:
		return true

	}
	return false
}
