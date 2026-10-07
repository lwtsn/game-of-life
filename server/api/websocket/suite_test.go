package websocket

import (
	"testing"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWebsocket(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Websocket Suite")
}

var _ = BeforeSuite(func() {
	gin.SetMode(gin.TestMode)
})
