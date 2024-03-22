package api

import (
	"github.com/HL/meta-bank/util"
	"github.com/go-playground/validator/v10"
)

var validateRole = func(fieldLevel validator.FieldLevel) bool {

	if role, ok := fieldLevel.Field().Interface().(string); ok {
		return util.IsSupportedRoles(role)
	}
	return false
}

var validateCurrency = func(fieldLevel validator.FieldLevel) bool {

	if currency, ok := fieldLevel.Field().Interface().(string); ok {
		return util.IsSupportedCurrency(currency)
	}
	return false
}
