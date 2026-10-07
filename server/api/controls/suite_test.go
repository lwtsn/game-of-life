package controls

import (
	"testing"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestControls(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controls Suite")
}

var _ = BeforeSuite(func() {
	gin.SetMode(gin.TestMode)
})
