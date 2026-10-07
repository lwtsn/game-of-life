package controls

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) Layout(c *gin.Context) {
	var body struct {
		Name string `json:"name"`
	}
	_ = c.ShouldBindJSON(&body)
	c.JSON(http.StatusOK, gin.H{"layout": body.Name, "applied": false})
}
