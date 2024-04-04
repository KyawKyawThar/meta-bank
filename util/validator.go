package util

const (
	USD  = "USD"
	EURO = "EURO"
	SGD  = "SGD"

	DEPOSITOR = "depositor" //As well as Users can only update their own information
	BANKER    = "banker"    // Banker on the other hand can update information of any depositor (users)
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
