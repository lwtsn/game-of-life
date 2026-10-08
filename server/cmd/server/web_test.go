package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("web root", func() {
	It("serves the built page without covering the socket", func() {
		dir := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(dir, "assets"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, "index.html"), []byte("home"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, "assets", "app.js"), []byte("js"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, "favicon.svg"), []byte("icon"), 0o644)).To(Succeed())

		gin.SetMode(gin.ReleaseMode)
		engine := gin.New()
		engine.GET("/ws", func(c *gin.Context) {
			c.String(http.StatusOK, "socket")
		})
		mountWeb(engine, dir)

		srv := httptest.NewServer(engine)
		DeferCleanup(srv.Close)

		get := func(path string) (int, string) {
			GinkgoHelper()
			res, err := http.Get(srv.URL + path)
			Expect(err).NotTo(HaveOccurred())
			defer res.Body.Close()
			body, err := io.ReadAll(res.Body)
			Expect(err).NotTo(HaveOccurred())
			return res.StatusCode, string(body)
		}

		status, body := get("/")
		Expect(status).To(Equal(http.StatusOK))
		Expect(body).To(Equal("home"))

		status, body = get("/assets/app.js")
		Expect(status).To(Equal(http.StatusOK))
		Expect(body).To(Equal("js"))

		status, body = get("/favicon.svg")
		Expect(status).To(Equal(http.StatusOK))
		Expect(body).To(Equal("icon"))

		status, body = get("/missing")
		Expect(status).To(Equal(http.StatusOK))
		Expect(body).To(Equal("home"))

		status, body = get("/ws")
		Expect(status).To(Equal(http.StatusOK))
		Expect(body).To(Equal("socket"))

		res, err := http.Post(srv.URL+"/nope", "text/plain", nil)
		Expect(err).NotTo(HaveOccurred())
		defer res.Body.Close()
		Expect(res.StatusCode).To(Equal(http.StatusNotFound))
	})
})
