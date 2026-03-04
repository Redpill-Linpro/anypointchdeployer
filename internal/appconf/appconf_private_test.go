package appconf

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Appconf private tests", func() {

	It("range unit test", func() {

		var currentVersion, desiredVersion string

		// Upgrade a patch version
		desiredVersion = "4.6.2"
		currentVersion = "4.6.1"
		Ω(runtimeVersionUpdated(desiredVersion, currentVersion)).Should(BeTrue(), fmt.Sprintf("%s -> %s should be true", currentVersion, desiredVersion))

		// Downgrade a patch version
		desiredVersion = "4.6.1"
		currentVersion = "4.6.2"
		Ω(runtimeVersionUpdated(desiredVersion, currentVersion)).Should(BeTrue(), fmt.Sprintf("%s -> %s should be true", currentVersion, desiredVersion))

		// Upgrade patch version validate ignore build and channel
		desiredVersion = "4.6.2"
		currentVersion = "4.6.1:3"
		Ω(runtimeVersionUpdated(desiredVersion, currentVersion)).Should(BeTrue(), fmt.Sprintf("%s -> %s should be true", currentVersion, desiredVersion))

		// Same version ignore build and channel
		desiredVersion = "4.6.1"
		currentVersion = "4.6.1:3"
		Ω(runtimeVersionUpdated(desiredVersion, currentVersion)).Should(BeFalse(), fmt.Sprintf("%s -> %s should be false", currentVersion, desiredVersion))

		// Downgrade patch version validate ignore build and channel
		desiredVersion = "4.6.10"
		currentVersion = "4.6.12:3"
		Ω(runtimeVersionUpdated(desiredVersion, currentVersion)).Should(BeTrue(), fmt.Sprintf("%s -> %s should be true", currentVersion, desiredVersion))
	})

	It("tilde range unit test", func() {

		var currentVersion, desiredVersion string

		// Same major and minor version, ignore patch level
		desiredVersion = "~4.6.1"
		currentVersion = "4.6.3"
		Ω(runtimeVersionUpdated(desiredVersion, currentVersion)).Should(BeFalse(), fmt.Sprintf("%s -> %s should be false", currentVersion, desiredVersion))

		// Same major and minor version, ignore patch level, but with build and channel
		desiredVersion = "~4.6.1"
		currentVersion = "4.6.13:3"
		Ω(runtimeVersionUpdated(desiredVersion, currentVersion)).Should(BeFalse(), fmt.Sprintf("%s -> %s should be false", currentVersion, desiredVersion))

		// Same major and minor version, double digit patch level, ignore patch level, but with build and channel
		desiredVersion = "~4.6.13"
		currentVersion = "4.6.23:3e"
		Ω(runtimeVersionUpdated(desiredVersion, currentVersion)).Should(BeFalse(), fmt.Sprintf("%s -> %s should be false", currentVersion, desiredVersion))

		// Same major and minor version, current patchlevel is earlier
		desiredVersion = "~4.6.13"
		currentVersion = "4.6.2:3"
		Ω(runtimeVersionUpdated(desiredVersion, currentVersion)).Should(BeTrue(), fmt.Sprintf("%s -> %s should be true", currentVersion, desiredVersion))

		// Same major and minor version, current patch level is later and double digit patch level
		desiredVersion = "~4.6.10"
		currentVersion = "4.6.12:33"
		Ω(runtimeVersionUpdated(desiredVersion, currentVersion)).Should(BeFalse(), fmt.Sprintf("%s -> %s should be false", currentVersion, desiredVersion))

		// Minor version different
		desiredVersion = "~4.6.10"
		currentVersion = "4.7.12:33"
		Ω(runtimeVersionUpdated(desiredVersion, currentVersion)).Should(BeTrue(), fmt.Sprintf("%s -> %s should be true", currentVersion, desiredVersion))

		// Major version different
		desiredVersion = "~4.6.10"
		currentVersion = "5.2.10:33"
		Ω(runtimeVersionUpdated(desiredVersion, currentVersion)).Should(BeTrue(), fmt.Sprintf("%s -> %s should be true", currentVersion, desiredVersion))
	})
})
