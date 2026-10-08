package life

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLife(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Life Suite")
}
