package flagvalidator

import (
	"fmt"

	"github.com/spf13/viper"
)

var validators = []SetValidator{}

type SetValidator interface {
	validateSet() error
}

type validFlagValueSet struct {
	flag        string
	validValues []any
}

func (fv validFlagValueSet) validateSet() error {
	for _, validValue := range fv.validValues {
		if viper.Get(fv.flag) == validValue {
			return nil
		}
	}
	return fmt.Errorf("Value '%s' is invalid for flag '%s'. Valid values "+
		"come from the set %v", viper.GetString(fv.flag), fv.flag, fv.validValues)
}

func AddFlagSetValidator(inputFlag string, inputValidValues []any) {
	validators = append(validators, validFlagValueSet{flag: inputFlag, validValues: inputValidValues})
}

func ValidateFlags() error {
	if resp := ValidateFlagSet(); resp != nil {
		return resp
	}
	if resp := validateAuthType(); resp != nil {
		return resp
	}
	return nil
}

func ValidateFlagSet() error {
	for _, val := range validators {
		if resp := val.validateSet(); resp != nil {
			return resp
		}
	}
	return nil
}

func validateAuthType() error {
	switch viper.GetString("authtype") {
	case "bearer":
		if !viper.IsSet("bearer") {
			return fmt.Errorf("Token must be supplied")
		}
	case "user":
		if !viper.IsSet("user") || !viper.IsSet("password") {
			return fmt.Errorf("User and password must be supplied")
		}
	case "connectedapp":
		if !viper.IsSet("client-id") || !viper.IsSet("client-secret") {
			return fmt.Errorf("Client id and secret must be supplied")
		}
	default:
		{
			return fmt.Errorf("Invalid authtype %s", viper.GetString("authtype"))
		}
	}
	return nil
}
