package flagvalidator

import (
	"testing"

	"github.com/spf13/viper"
)

func TestValidateFlagSet(t *testing.T) {

	AddFlagSetValidator("region", []any{"US", "EU"})
	AddFlagSetValidator("authType", []any{"bearer", "user", "connectedapp"})
	AddFlagSetValidator("concurrent-deployments", []any{1, 2, 3, 4, 5})

	viper.Set("region", "US")
	viper.Set("authType", "user")
	viper.Set("concurrent-deployments", 5)

	resp := ValidateFlagSet()
	if resp != nil {
		t.Errorf("%s", resp)
	}
}

func TestValidateAuthType(t *testing.T) {
	viper.Set("authtype", "user")
	viper.Set("user", "someuser")
	viper.Set("password", "somepassword")

	if resp := validateAuthType(); resp != nil {
		t.Errorf("%s", resp)
	}

	viper.Reset()
	viper.Set("authtype", "connectedapp")
	viper.Set("client-id", "some-id")
	viper.Set("client-secret", "some-secret")

	if resp := validateAuthType(); resp != nil {
		t.Errorf("%s", resp)
	}

	viper.Reset()
	viper.Set("authtype", "bearer")
	viper.Set("bearer", "some-bearer")

	if resp := validateAuthType(); resp != nil {
		t.Errorf("%s", resp)
	}

	viper.Reset()
	viper.Set("authtype", "bad-authtype")

	if resp := validateAuthType(); resp == nil {
		t.Errorf("%s", resp)
	}

}
