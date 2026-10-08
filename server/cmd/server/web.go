package main

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// mountWeb serves a Vite build. Registered routes such as /ws stay in front of this.
func mountWeb(engine *gin.Engine, root string) {
	root = filepath.Clean(root)
	engine.NoRoute(func(c *gin.Context) {
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.Status(http.StatusNotFound)
			return
		}
		rel := strings.TrimPrefix(path.Clean("/"+c.Request.URL.Path), "/")
		if rel != "" {
			candidate := filepath.Join(root, filepath.FromSlash(rel))
			if fileInside(root, candidate) {
				info, err := os.Stat(candidate)
				if err == nil && info.Mode().IsRegular() {
					c.File(candidate)
					return
				}
			}
		}
		c.File(filepath.Join(root, "index.html"))
	})
}

func fileInside(root, candidate string) bool {
	rel, err := filepath.Rel(root, candidate)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}
