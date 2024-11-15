package api

import (
	"errors"
	"fmt"
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
	ErrorInSyntax                  = &pgconn.PgError{Code: "42601"}
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

	errorResponse := map[string]struct {
		message string
		Code    int
	}{
		ErrorUniqueViolation.Code:           {"A record with the same unique identifier already exists.", http.StatusForbidden},
		ErrForeignViolation.Code:            {"The operation cannot be completed because it would violate a foreign key constraint.", http.StatusForbidden},
		ErrNotNullViolation.Code:            {"A required column is missing a value.", http.StatusBadRequest}, // Not null violation
		ErrorCheckViolation.Code:            {"The value violates a constraint.", http.StatusBadRequest},      // Check violation
		ErrorDeadlockDetected.Code:          {"A deadlock was detected while trying to complete your request. Please try again.", http.StatusConflict},
		ErrorStringDataRightTruncation.Code: {"The data exceeds the maximum allowed length.", http.StatusBadRequest}, // String data right truncation
		ErrorNumericValueOutOfRange.Code:    {"The numeric value is out of range.", http.StatusBadRequest},           // Numeric value out of range
		ErrorInvalidTextRepresentation.Code: {"The text representation of a value is invalid.", http.StatusBadRequest},
		ErrorExclusionViolation.Code:        {"The operation violates an exclusion constraint.", http.StatusForbidden},
		ErrConnectionDone.Code:              {"The database connection is done.", http.StatusInternalServerError},
		ErrorInSyntax.Code:                  {"There was a syntax error in the SQL statement.", http.StatusBadRequest}, //
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {

		if response, found := errorResponse[pgErr.Code]; found {
			return response.message, response.Code
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
		messages := make([]string, 0, len(validationError))

		for _, fieldErr := range validationError {

			messages = append(messages, formatValidationError(fieldErr))

			//fieldName := fieldErr.Field()
			//switch fieldName {
			//case "Username":
			//	messages = append(messages, formatValidationError(fieldName, fieldErr))
			//	break
			//case "Password":
			//	messages = append(messages, formatValidationError(fieldName, fieldErr))
			//	break
			//case "Role":
			//	messages = append(messages, formatValidationError(fieldName, fieldErr))
			//	break
			//case "Email":
			//	messages = append(messages, formatValidationError(fieldName, fieldErr))
			//	break
			//case "FullName":
			//	messages = append(messages, formatValidationError(fieldName, fieldErr))
			//	break
			//
			//case "Balance":
			//	messages = append(messages, formatValidationError(fieldName, fieldErr))
			//	break
			//case "Currency":
			//	messages = append(messages, formatValidationError(fieldName, fieldErr))
			//	break
			//
			//case "ID":
			//	messages = append(messages, formatValidationError(fieldName, fieldErr))
			//	break
			//
			//case "ReceiveAccountID":
			//	messages = append(messages, formatValidationError(fieldName, fieldErr))
			//	break
			//case "TransferAccountID":
			//	messages = append(messages, formatValidationError(fieldName, fieldErr))
			//	break
			//case "Amount":
			//	messages = append(messages, formatValidationError(fieldName, fieldErr))
			//	break
			//case "EmailId":
			//	messages = append(messages, formatValidationError(fieldName, fieldErr))
			//	break
			//case "SecretCode":
			//	messages = append(messages, formatValidationError(fieldName, fieldErr))
			//	break
			//}
		}
		//if len(messages) > 0 {
		//	return strings.Join(messages, "\n") // Join multiple messages with newlines
		//}

	}

	return strings.Split(err.Error(), ":")[1] //Default validation message

}

// formatValidationError is a function to format user validation error
func formatValidationError(err validator.FieldError) string {
	// Define default messages for validation tags to reduce repetitive code
	validationMessages := map[string]string{
		"required": "%s is required.",
		"alphanum": "%s must contain only letters and numbers.",
		"email":    "%s is not a valid email address.",
		"currency": "currency must be SGD, USD, or EURO.",
		"role":     "roles must be BANKER or DEPOSITOR.",
		"min":      "%s must be at least %s.",
		"max":      "%s must not exceed %s.",
		"gt":       "%s must be greater than %s.",
		"gte":      "%s must be greater than or equal to %s.",
		"lt":       "%s must be less than %s.",
		"lte":      "%s must be less than or equal to %s.",
		"uuid":     "%s must be a valid UUID.",
		"eqfield":  "%s must be equal to %s.",
	}

	field := err.Field()
	tag := err.Tag()
	param := err.Param()
	if msg, found := validationMessages[tag]; found {
		if tag == "min" || tag == "max" || tag == "gt" || tag == "gte" || tag == "lt" || tag == "lte" || tag == "eqfield" {
			return fmt.Sprintf(msg, field, param) // Include param for tags that require it
		}
		return fmt.Sprintf(msg, field)
	}

	return fmt.Sprintf("Validation error for %s: %v", field, err.Error()) // Fallback for unknown tags

}

//func formatValidationError(field string, err validator.FieldError) string {
//
//	switch err.Tag() {
//	case "required":
//		return fmt.Sprintf("%s is required.", field)
//	case "alphanum":
//		return fmt.Sprintf("%s must contain only letters and numbers.", field)
//	case "min":
//		return fmt.Sprintf("%s must be at least %s", field, err.Param())
//	case "max":
//		return fmt.Sprintf("%s must be at least %s", field, err.Param())
//	case "gt":
//		return fmt.Sprintf("%s must not be negative", field)
//	case "email":
//		return fmt.Sprintf("%s is not a valid email address.", field)
//	case "currency":
//		return fmt.Sprintf("currency must be %s, %s OR %s", util.SGD, util.USD, util.EURO)
//	case "role":
//		return fmt.Sprintf("roles must be %s OR %s", util.BANKER, util.DEPOSITOR)
//
//	default:
//		return fmt.Sprintf("Validation error for %s: %v", field, err.Error()) // Generic message for unknown tags
//
//	}
//}
