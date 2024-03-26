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

	fmt.Printf("ERROR %v", err.Error())
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		fmt.Printf("pgError %v", pgErr)
		switch {
		case errors.Is(pgErr, ErrorUniqueViolation): // UniqueViolation in pgconn (check actual code for your driver)
			return "A record with the same unique identifier already exists.", http.StatusForbidden
		case errors.Is(pgErr, ErrForeignViolation): // ForeignKeyViolation in pgconn (check actual code for your driver)
			return "The operation cannot be completed because it would violate a foreign key constraint.", http.StatusForbidden
		case errors.Is(pgErr, ErrNotNullViolation): // NotNullViolation in pgconn (check actual code for your driver)
			return "One or more required fields are missing a value.", http.StatusBadRequest // Customize based on your schema
		case errors.Is(pgErr, ErrorCheckViolation): // NotNullViolation in pgconn (check actual code for your driver)
			return "The data you provided doesn't meet the required criteria.", http.StatusBadRequest
		case errors.Is(pgErr, ErrorRecordNotFound):
			return "The requested record was not found.", http.StatusNotFound
		default:
			return fmt.Sprintf("pgx error: %v (%w)", pgErr.Message, pgErr), http.StatusInternalServerError // Use Message for error message, include original error
		}
	} else {
		// Handle other errors (not pgx-specific)
		return "An unexpected database error occurred.", http.StatusInternalServerError // Adjust message as needed
	}

}

// GetMessageFromUserValidationError is a function for user validation fields
func GetMessageFromUserValidationError(err error) string {
	// Type assertion on errors
	//if validationError, ok := err.(validator.ValidationErrors); ok {}
	fmt.Println("error is:", err)
	var validationError validator.ValidationErrors
	if errors.As(err, &validationError) {
		var messages []string
		for _, fieldErr := range validationError {
			fieldName := fieldErr.Field()

			switch fieldName {
			case "Username":
				messages = append(messages, formatUserValidationError(fieldName, fieldErr))
				break
			case "Password":
				messages = append(messages, formatUserValidationError(fieldName, fieldErr))
				break
			case "Role":
				messages = append(messages, formatUserValidationError(fieldName, fieldErr))
				break
			case "Email":
				messages = append(messages, formatUserValidationError(fieldName, fieldErr))
				break
			case "FullName":
				messages = append(messages, formatUserValidationError(fieldName, fieldErr))
				break
			case "IsActive":
				messages = append(messages, formatUserValidationError(fieldName, fieldErr))

			}
		}
		if len(messages) > 0 {
			return strings.Join(messages, "\n") // Join multiple messages with newlines
		}

	}

	return "An error occurred during validation." //Default validation message

}

// formatUserValidationError is a function to format user validation error
func formatUserValidationError(field string, err validator.FieldError) string {

	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required.", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers.", field)
	case "min":
		return fmt.Sprintf("%s must be at least %d characters long.", field, err.Param())
	case "email":
		return fmt.Sprintf("%s is not a valid email address.", field)
	case "role":
		return fmt.Sprintf("roles must be %s OR %s", util.ADMIN, util.USER)

	default:
		return fmt.Sprintf("Validation error for %s: %v", field, err.Error()) // Generic message for unknown tags

	}
}
