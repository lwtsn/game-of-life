package life

import (
	"testing"

	lifepb "game_of_life/server/gen/life/v1"
	"game_of_life/server/internal/grid/source"

	"go.uber.org/fx"
)

func BenchmarkNext80x50(b *testing.B) {
	benchmarkNext(b, int(lifepb.GridSize_GRID_SIZE_WIDTH), int(lifepb.GridSize_GRID_SIZE_HEIGHT))
}

func BenchmarkNextMillion(b *testing.B) {
	benchmarkNext(b, 1000, 1000)
}

func benchmarkNext(b *testing.B, width, height int) {
	b.Helper()
	var src source.Source
	app := fx.New(Module, fx.Populate(&src), fx.NopLogger)
	if err := app.Err(); err != nil {
		b.Fatal(err)
	}
	cells := make([]source.Cell, width*height)
	for i := range cells {
		if i%7 == 0 {
			cells[i] = source.Cell{Alive: true}
		}
	}
	current := snapshot{width: width, height: height, cells: cells}
	b.ReportAllocs()
	for b.Loop() {
		src.Next(current)
	}
}
