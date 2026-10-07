package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var body struct {
	Name string `json:"name"`
}

func (h *hub) layout(c *gin.Context) {

	_ = c.ShouldBindJSON(&body)
	c.JSON(http.StatusOK, gin.H{"layout": body.Name, "applied": false})
}
