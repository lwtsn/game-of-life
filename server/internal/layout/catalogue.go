package layout

import (
	"embed"
	"fmt"
	"slices"

	lifepb "game_of_life/server/gen/life/v1"

	"google.golang.org/protobuf/encoding/prototext"
)

//go:embed patterns.textproto
var rawCatalogue embed.FS

// Cell is one square of a shape, relative to the shape's top-left.
type Cell struct {
	X int
	Y int
}

// Shape is a pattern the server can stamp.
type Shape struct {
	Pattern lifepb.Pattern
	Label   string
	Cells   []Cell
	Width   int
	Height  int
}

// Catalogue is the set of shapes the server can stamp.
type Catalogue interface {
	// ByPattern looks up one shape.
	ByPattern(pattern lifepb.Pattern) (Shape, bool)
	// Shapes returns every shape in the order patterns.textproto lists them.
	Shapes() []Shape
}

type catalogue struct {
	shapes map[lifepb.Pattern]Shape
	order  []Shape
}

func (c catalogue) ByPattern(pattern lifepb.Pattern) (Shape, bool) {
	shape, ok := c.shapes[pattern]
	return shape, ok
}

// Shapes returns a copy, so a caller cannot reorder the catalogue.
func (c catalogue) Shapes() []Shape {
	return slices.Clone(c.order)
}

func loadCatalogue() (Catalogue, error) {
	raw, err := rawCatalogue.ReadFile("patterns.textproto")
	if err != nil {
		return nil, err
	}
	var decoded lifepb.Catalogue
	if err := prototext.Unmarshal(raw, &decoded); err != nil {
		return nil, err
	}
	required := []lifepb.Pattern{
		lifepb.Pattern_PATTERN_BLOCK,
		lifepb.Pattern_PATTERN_BLINKER,
		lifepb.Pattern_PATTERN_GLIDER,
		lifepb.Pattern_PATTERN_BEACON,
	}
	shapes := make(map[lifepb.Pattern]Shape, len(required))
	order := make([]Shape, 0, len(required))
	for _, item := range decoded.GetShapes() {
		pattern := item.GetPattern()
		if pattern == lifepb.Pattern_PATTERN_UNSPECIFIED {
			return nil, fmt.Errorf("catalogue shape has no pattern")
		}
		if _, exists := shapes[pattern]; exists {
			return nil, fmt.Errorf("catalogue repeats %s", pattern)
		}
		shape, err := shapeFrom(item)
		if err != nil {
			return nil, err
		}
		shapes[pattern] = shape
		order = append(order, shape)
	}
	for _, pattern := range required {
		if _, ok := shapes[pattern]; !ok {
			return nil, fmt.Errorf("catalogue is missing %s", pattern)
		}
	}
	if len(shapes) != len(required) {
		return nil, fmt.Errorf("catalogue has %d shapes", len(shapes))
	}
	return catalogue{shapes: shapes, order: order}, nil
}

func shapeFrom(item *lifepb.Shape) (Shape, error) {
	pattern := item.GetPattern()
	if item.GetLabel() == "" {
		return Shape{}, fmt.Errorf("catalogue shape %s has no label", pattern)
	}
	if len(item.GetCells()) == 0 {
		return Shape{}, fmt.Errorf("catalogue shape %s has no cells", pattern)
	}
	cells := make([]Cell, 0, len(item.GetCells()))
	maxX, maxY := 0, 0
	for _, offset := range item.GetCells() {
		x := int(offset.GetX())
		y := int(offset.GetY())
		if x < 0 || y < 0 {
			return Shape{}, fmt.Errorf("catalogue shape %s has a negative cell", pattern)
		}
		if x > maxX {
			maxX = x
		}
		if y > maxY {
			maxY = y
		}
		cells = append(cells, Cell{X: x, Y: y})
	}
	return Shape{
		Pattern: pattern,
		Label:   item.GetLabel(),
		Cells:   cells,
		Width:   maxX + 1,
		Height:  maxY + 1,
	}, nil
}
