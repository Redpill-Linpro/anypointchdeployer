package anypointclient

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Login", func() {
	It("should be successfull", func() {
		fixture, err := ioutil.ReadFile("testdata//login/successfull-login-response.json")
		if err != nil {
			Fail(fmt.Sprintf("Failed %v", err))
		}
		httpmock.RegisterResponder("POST", "/accounts/login", func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(200, string(fixture))
			resp.Header.Add("Content-Type", "application/json")
			return resp, nil
		})

		err = client.Login()
		Ω(err == nil).Should(BeTrue(), "Error is %+v", err)
		Ω(client.username).Should(Equal("user"), "username")
		Ω(client.bearer).Should(Equal("12345678-1234-1234-1234-123456789101"), "bearer")
	})
})
