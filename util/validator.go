package util

const (
	USD  = "USD"
	EURO = "EURO"
	SGD  = "SGD"
	JPY  = "JPY"
	GBP  = "GBP"
	AUD  = "AUD"
	CAD  = "CAD"
	CHF  = "CHF"
	NZD  = "NZD"
	SEK  = "SEK"
	NOK  = "NOK"
	DKK  = "DKK"

	DEPOSITOR = "depositor" //As well as Users can only update their own information
	BANKER    = "banker"    // Banker on the other hand can update information of any depositor (users)
)

func IsSupportedCurrency(currency string) bool {

	switch currency {
	case USD, EURO, SGD, JPY, GBP, AUD, CAD, CHF, NZD, SEK, NOK, DKK:
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
