package layout

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLayoutAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Layout API Suite")
}
