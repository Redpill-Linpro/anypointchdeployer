package anypointclient

import (
	"io/ioutil"
	"net/http"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Environment", func() {
	It("should find the environment", func() {
		fixture, err := ioutil.ReadFile("testdata/environment/environments-response.json")
		httpmock.RegisterResponder("GET", "/accounts/api/organizations/12345678-6085-4179-9bed-917f6643df29/environments", func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(200, string(fixture))
			resp.Header.Add("Content-Type", "application/json")
			return resp, nil
		})

		env, err := client.ResolveEnvironment(Organization{
			ID:   "12345678-6085-4179-9bed-917f6643df29",
			Name: "Example Inc",
		}, "Sandbox")
		Ω(err == nil).Should(BeTrue(), "Error is %v", err)
		Ω(env.ID).Should(Equal("12345678-1707-4beb-8142-1899dd37a3df"), "environment ID")
	})
})
