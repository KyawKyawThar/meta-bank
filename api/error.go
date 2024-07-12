package api

import (
	"errors"
	"fmt"
	"github.com/HL/meta-bank/util"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"net/http"
	"strings"
)

var (
	//Common errors
	ErrorRecordNotFound  = pgx.ErrNoRows
	ErrorUniqueViolation = &pgconn.PgError{Code: "23505"}
	ErrForeignViolation  = &pgconn.PgError{Code: "23503"}
	ErrNotNullViolation  = &pgconn.PgError{Code: "23502"}
	ErrorCheckViolation  = &pgconn.PgError{Code: "23514"}
	ErrConnectionDone    = &pgconn.PgError{Code: "08003"}

	// Additional errors
	ErrorDeadlockDetected          = &pgconn.PgError{Code: "40P01"}
	ErrorStringDataRightTruncation = &pgconn.PgError{Code: "22001"}
	ErrorNumericValueOutOfRange    = &pgconn.PgError{Code: "22003"}
	ErrorInvalidTextRepresentation = &pgconn.PgError{Code: "22P02"}
	ErrorExclusionViolation        = &pgconn.PgError{Code: "23P01"}
)

// CheckError verifies if an error matches a specific type, returning a descriptive message or nil
func CheckError(err error, matchers ...func(err error) bool) error {

	fmt.Println("matcher from CheckError", matchers)

	for _, matcher := range matchers {
		if matcher(err) {
			return err
		}

	}
	return nil
}

// GetMessageFromDBError to extract a human-readable message from a database error
func GetMessageFromDBError(err error) (string, int) {

	if errors.Is(err, ErrorRecordNotFound) {
		return "The requested record was not found.", http.StatusNotFound
	}

	//if pgErr, ok := err.(*pgconn.PgError); ok {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case ErrorUniqueViolation.Code: // duplicate value is being inserted for a column or set of columns that have a unique constraint defined.
			return "A record with the same unique identifier already exists.", http.StatusForbidden
		case ErrForeignViolation.Code:
			return "The operation cannot be completed because it would violate a foreign key constraint.", http.StatusForbidden
		case ErrNotNullViolation.Code:
			return "One or more required fields are missing a value.", http.StatusBadRequest
		case ErrorCheckViolation.Code: // This constraint enforces certain conditions on data values for integrity and consistency.
			return "The data you provided doesn't meet the required criteria.", http.StatusBadRequest
		case ErrorDeadlockDetected.Code:
			return "A deadlock was detected while trying to complete your request. Please try again.", http.StatusConflict
		case ErrorStringDataRightTruncation.Code:
			return "The data provided is too long for the designated field.", http.StatusBadRequest
		case ErrorNumericValueOutOfRange.Code:
			return "The numeric value provided is out of range.", http.StatusBadRequest
		case ErrorInvalidTextRepresentation.Code:
			return "The text representation of a value is invalid.", http.StatusBadRequest
		case ErrorExclusionViolation.Code:
			return "The operation violates an exclusion constraint.", http.StatusForbidden
		case ErrConnectionDone.Code:
			return "The database connection is done.", http.StatusInternalServerError
		}
	}
	return err.Error(), http.StatusInternalServerError

}

// GetMessageFromUserValidationError is a function for user validation fields
func GetMessageFromUserValidationError(err error) string {
	// Type assertion on errors
	//if validationError, ok := err.(validator.ValidationErrors); ok {}
	var validationError validator.ValidationErrors
	fmt.Println("GetMessageFromUserValidationError", err)
	if errors.As(err, &validationError) {
		var messages []string
		for _, fieldErr := range validationError {

			fieldName := fieldErr.Field()

			switch fieldName {
			case "Username":
				messages = append(messages, formatValidationError(fieldName, fieldErr))
				break
			case "Password":
				messages = append(messages, formatValidationError(fieldName, fieldErr))
				break
			case "Role":
				messages = append(messages, formatValidationError(fieldName, fieldErr))
				break
			case "Email":
				messages = append(messages, formatValidationError(fieldName, fieldErr))
				break
			case "FullName":
				messages = append(messages, formatValidationError(fieldName, fieldErr))
				break

			case "Balance":
				messages = append(messages, formatValidationError(fieldName, fieldErr))
				break
			case "Currency":
				messages = append(messages, formatValidationError(fieldName, fieldErr))
				break

			case "ID":
				messages = append(messages, formatValidationError(fieldName, fieldErr))
				break

			case "ReceiveAccountID":
				messages = append(messages, formatValidationError(fieldName, fieldErr))
				break
			case "TransferAccountID":
				messages = append(messages, formatValidationError(fieldName, fieldErr))
				break
			case "Amount":
				messages = append(messages, formatValidationError(fieldName, fieldErr))
				break

			}
		}
		if len(messages) > 0 {
			return strings.Join(messages, "\n") // Join multiple messages with newlines
		}

	}

	errString := strings.Split(err.Error(), ":")[1]

	return errString //Default validation message

}

// formatValidationError is a function to format user validation error
func formatValidationError(field string, err validator.FieldError) string {

	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required.", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers.", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, err.Param())
	case "max":
		return fmt.Sprintf("%s must be at least %s", field, err.Param())
	case "gt":
		return fmt.Sprintf("%s must not be negative", field)
	case "email":
		return fmt.Sprintf("%s is not a valid email address.", field)
	case "currency":
		return fmt.Sprintf("currency must be %s, %s OR %s", util.SGD, util.USD, util.EURO)
	case "role":
		return fmt.Sprintf("roles must be %s OR %s", util.BANKER, util.DEPOSITOR)

	default:
		return fmt.Sprintf("Validation error for %s: %v", field, err.Error()) // Generic message for unknown tags

	}
}
