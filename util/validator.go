package util

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
)

const (
	USD  = "USD"
	EURO = "EURO"
	SGD  = "SGD"

	DEPOSITOR = "depositor" //As well as Users can only update their own information
	BANKER    = "banker"    // Banker on the other hand can update information of any depositor (users)

)

var (
	isValidUsername = regexp.MustCompile("^[a-z-Z0-9_]+$").MatchString
	isValidFullName = regexp.MustCompile("^[a-zA-Z\\s]+$").MatchString
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

func validateString(value string, minLength int, maxLength int) error {

	n := len(value)

	if n < minLength || n > maxLength {
		return fmt.Errorf("must container between %d and %d characters", minLength, maxLength)
	}
	return nil
}

func ValidateUsername(username string) error {

	if err := validateString(username, 3, 100); err != nil {
		return err
	}

	if !isValidUsername(username) {
		return errors.New("must contain only lowercase letters, digits, or underscore")
	}

	return nil
}

func ValidateFullName(fullName string) error {
	if err := validateString(fullName, 3, 100); err != nil {
		return err
	}

	if !isValidFullName(fullName) {
		return fmt.Errorf("must contain only letters and space")
	}
	return nil
}

func ValidatePassword(password string) error {

	return validateString(password, 3, 100)
}

func ValidateEmail(email string) error {

	fmt.Println("code is running in here..", email)
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("invalid email address")
	}

	return nil
}

func ValidateEmailId(id int64) error {
	if id <= 0 {
		return fmt.Errorf("must be a positive integer")

	}
	return nil
}

func ValidateSecretCode(value string) error {
	return validateString(value, 32, 128)
}
