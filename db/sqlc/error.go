package db

import (
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	//Common errors
	ErrorRecordNotFound  = pgx.ErrNoRows
	ErrorUniqueViolation = &pgconn.PgError{Code: "23505"}
	ErrForeignViolation  = &pgconn.PgError{Code: "23503"}
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

// Error types for common database errors

func IsUniqueViolation(err error) bool {
	return errors.Is(err, ErrorUniqueViolation)

}

func IsForeignViolation(err error) bool {
	return errors.Is(err, ErrForeignViolation)
}

// GetMessageFromDBError to extract a human-readable message from a database error
func GetMessageFromDBError(err error) string {

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		return pgErr.Message
	}
	return err.Error()
}

//TODO: Move Below Function To Server.go file
//func handleDBError(c *gin.Context, err error) {
//    statusCode := http.StatusInternalServerError
//
//    if CheckError(err, isUniqueViolation, isForeignViolation) != nil {
//        // Handle specific database errors with custom messages
//        statusCode = http.StatusConflict
//        c.JSON(statusCode, gin.H{"error": GetMessageFromDBError(err)})
//    } else {
//        // Handle generic database errors
//        statusCode = http.StatusInternalServerError
//        c.JSON(statusCode, gin.H{"error": GetMessageFromDBError(err)})
//    }
//}
