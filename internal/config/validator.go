package config

import "github.com/go-playground/validator/v10"

var Validate = validator.New()

func ValidatorInit() {
	// Validation for numeric string
	Validate.RegisterValidation("numericstr", func(fl validator.FieldLevel) bool {
		str := fl.Field().String()
		for _, ch := range str {
			if ch < '0' || ch > '9' {
				return false
			}
		}

		return true
	})
}
