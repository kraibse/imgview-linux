package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
)

type mockImage struct {
	bounds image.Rectangle
}

func (m *mockImage) ColorModel() color.Model { return color.RGBAModel }
func (m *mockImage) Bounds() image.Rectangle { return m.bounds }
func (m *mockImage) At(x, y int) color.Color { return color.RGBA{0, 0, 0, 255} }

func newMockImage(w, h int) *mockImage {
	return &mockImage{bounds: image.Rect(0, 0, w, h)}
}

func TestPtInSel(t *testing.T) {
	g := &Game{
		hasSel:    true,
		selStartX: 10,
		selStartY: 10,
		selEndX:   50,
		selEndY:   50,
	}

	cases := []struct {
		px, py float64
		want   bool
	}{
		{30, 30, true},
		{10, 10, true},
		{50, 50, true},
		{5, 5, false},
		{55, 55, false},
		{30, 5, false},
		{5, 30, false},
	}

	for _, c := range cases {
		got := g.ptInSel(c.px, c.py)
		if got != c.want {
			t.Errorf("ptInSel(%v, %v) = %v, want %v", c.px, c.py, got, c.want)
		}
	}
}

func TestPtInSelNoSelection(t *testing.T) {
	g := &Game{hasSel: false}
	if g.ptInSel(10, 10) {
		t.Error("ptInSel should return false when no selection exists")
	}
}

func TestScreenToImg(t *testing.T) {
	g := &Game{
		img:     newMockImage(100, 100),
		scale:   2.0,
		offsetX: 10,
		offsetY: 20,
	}

	sx, sy := float64(120), float64(140)
	ww, wh := 200, 200
	ix, iy := g.screenToImg(sx, sy, ww, wh)

	dw := 100 * 2.0
	dh := 100 * 2.0
	expectedIX := (float64(ww)-dw)/2 + g.offsetX
	expectedIY := (float64(wh)-dh)/2 + g.offsetY
	wantX := (sx - expectedIX) / g.scale
	wantY := (sy - expectedIY) / g.scale

	if ix != wantX || iy != wantY {
		t.Errorf("screenToImg(%v, %v) = (%v, %v), want (%v, %v)", sx, sy, ix, iy, wantX, wantY)
	}
}

func TestSelImgCoords(t *testing.T) {
	g := &Game{
		img:       newMockImage(200, 200),
		scale:     1.0,
		offsetX:   0,
		offsetY:   0,
		selStartX: 50,
		selStartY: 50,
		selEndX:   150,
		selEndY:   150,
	}

	minX, minY, maxX, maxY := g.selImgCoords(200, 200)
	if minX < 0 || minY < 0 || maxX > 200 || maxY > 200 {
		t.Errorf("selImgCoords out of bounds: (%d, %d, %d, %d)", minX, minY, maxX, maxY)
	}
	if minX >= maxX || minY >= maxY {
		t.Errorf("selImgCoords invalid rect: (%d, %d, %d, %d)", minX, minY, maxX, maxY)
	}
}

func TestSelImgCoordsClamping(t *testing.T) {
	g := &Game{
		img:       newMockImage(100, 100),
		scale:     1.0,
		offsetX:   0,
		offsetY:   0,
		selStartX: -50,
		selStartY: -50,
		selEndX:   200,
		selEndY:   200,
	}

	minX, minY, maxX, maxY := g.selImgCoords(100, 100)
	if minX != 0 || minY != 0 || maxX != 100 || maxY != 100 {
		t.Errorf("selImgCoords clamping failed: (%d, %d, %d, %d)", minX, minY, maxX, maxY)
	}
}

func TestLoadImage(t *testing.T) {
	path := "testdata/test_image.png"
	if err := os.MkdirAll("testdata", 0755); err != nil {
		t.Fatalf("failed to create testdata: %v", err)
	}
	if err := createTestImage(path, 50, 50); err != nil {
		t.Fatalf("failed to create test image: %v", err)
	}
	defer os.Remove(path)

	g := &Game{scale: 1.0}
	g.loadImage(path)
	if !g.hasImg {
		t.Fatal("loadImage should set hasImg to true")
	}
	if g.img == nil {
		t.Fatal("loadImage should set img")
	}
	b := g.img.Bounds()
	if b.Dx() != 50 || b.Dy() != 50 {
		t.Errorf("loaded image size = %dx%d, want 50x50", b.Dx(), b.Dy())
	}
}

func TestLoadImageNotFound(t *testing.T) {
	g := &Game{scale: 1.0}
	g.loadImage("testdata/nonexistent.png")
	if g.hasImg {
		t.Error("loadImage should not set hasImg for missing file")
	}
}

func TestMathHelpers(t *testing.T) {
	if min(1, 2) != 1 {
		t.Error("min(1, 2) != 1")
	}
	if max(1, 2) != 2 {
		t.Error("max(1, 2) != 2")
	}
	if abs(-5) != 5 {
		t.Error("abs(-5) != 5")
	}
	if abs(5) != 5 {
		t.Error("abs(5) != 5")
	}
}

func createTestImage(path string, w, h int) error {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
