package grid

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGrid(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Grid Suite")
}
