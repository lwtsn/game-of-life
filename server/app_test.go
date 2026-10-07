package main

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	"go.uber.org/fx"
)

func TestAppStarts(t *testing.T) {
	g := NewWithT(t)

	app := fx.New(
		module(),
		fx.Replace(Config{Addr: "127.0.0.1:0"}),
		fx.NopLogger,
	)
	g.Expect(app.Err()).NotTo(HaveOccurred())

	ctx := context.Background()
	g.Expect(app.Start(ctx)).To(Succeed())
	g.Expect(app.Stop(ctx)).To(Succeed())
}
