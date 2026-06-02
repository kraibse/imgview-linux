package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.design/x/clipboard"
	"golang.org/x/image/bmp"
	"golang.org/x/image/font/basicfont"
	_ "golang.org/x/image/webp"
)

type Game struct {
	img       *ebiten.Image
	hasImg    bool
	scale     float64
	offsetX   float64
	offsetY   float64
	selecting bool
	selStartX float64
	selStartY float64
	selEndX   float64
	selEndY   float64
	hasSel    bool
	menuOpen  bool
	menuX     int
	menuY     int
	panning   bool
	panStartX float64
	panStartY float64
	panOffX   float64
	panOffY   float64
	dragWin   bool
	dragSX    int
	dragSY    int
	winSX     int
	winSY     int
	hoverCtrl bool
	closeHov  bool
	maxHov    bool
	minHov    bool
	prevLeft  bool
	prevRight bool
	prevMid   bool
	prevCtrlS bool
}

func (g *Game) Update() error {
	if !g.hasImg {
		return nil
	}

	cx, cy := ebiten.CursorPosition()
	mx, my := float64(cx), float64(cy)
	ww, _ := ebiten.WindowSize()

	left := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	mid := ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle)
	right := ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)

	g.hoverCtrl = cy >= 0 && cy <= 40 && cx >= ww-100 && cx <= ww
	g.minHov = cx >= ww-98 && cx <= ww-78 && cy >= 15 && cy <= 35
	g.maxHov = cx >= ww-68 && cx <= ww-48 && cy >= 15 && cy <= 35
	g.closeHov = cx >= ww-38 && cx <= ww-18 && cy >= 15 && cy <= 35

	_, wheel := ebiten.Wheel()
	if wheel != 0 && !g.menuOpen {
		g.scale *= 1.0 + wheel*0.1
		if g.scale < 0.1 {
			g.scale = 0.1
		}
		if g.scale > 20 {
			g.scale = 20
		}
	}

	if left && !g.prevLeft && !g.selecting && !g.dragWin && cy >= 0 && cy <= 40 && !g.closeHov && !g.maxHov && !g.minHov {
		g.dragWin = true
		g.dragSX, g.dragSY = cx, cy
		g.winSX, g.winSY = ebiten.WindowPosition()
	}
	if g.dragWin {
		if !left {
			g.dragWin = false
		} else {
			newX := g.winSX + (cx - g.dragSX)
			newY := g.winSY + (cy - g.dragSY)
			ebiten.SetWindowPosition(newX, newY)
			g.winSX, g.winSY = newX, newY
			g.dragSX, g.dragSY = cx, cy
		}
	}

	if mid && !g.prevMid && !g.panning {
		g.panning = true
		g.panStartX, g.panStartY = mx, my
		g.panOffX, g.panOffY = g.offsetX, g.offsetY
	}
	if g.panning {
		if !mid {
			g.panning = false
		} else {
			g.offsetX = g.panOffX + (mx - g.panStartX)
			g.offsetY = g.panOffY + (my - g.panStartY)
		}
	}

	if left && !g.prevLeft && !g.dragWin && !g.panning && cy > 40 && !g.menuOpen {
		g.selecting = true
		g.selStartX, g.selStartY = mx, my
		g.selEndX, g.selEndY = mx, my
		g.hasSel = false
	}
	if g.selecting {
		if !left {
			g.selecting = false
			if abs(g.selEndX-g.selStartX) > 2 && abs(g.selEndY-g.selStartY) > 2 {
				g.hasSel = true
			}
		} else {
			g.selEndX, g.selEndY = mx, my
		}
	}

	if left && !g.prevLeft && g.menuOpen {
		if cx >= g.menuX && cx <= g.menuX+140 && cy >= g.menuY && cy <= g.menuY+105 {
			switch (cy - g.menuY) / 35 {
			case 0:
				g.crop()
			case 1:
				g.copy()
			case 2:
				g.saveImage()
			}
		}
		g.menuOpen = false
	}

	if right && !g.prevRight {
		if g.hasSel && g.ptInSel(mx, my) {
			g.menuOpen = true
			g.menuX, g.menuY = cx, cy
		} else {
			g.menuOpen = false
			g.hasSel = false
		}
	}

	if left && !g.prevLeft && g.hoverCtrl {
		if g.closeHov {
			os.Exit(0)
		}
		if g.maxHov {
			cw, ch := ebiten.WindowSize()
			mw, mh := ebiten.Monitor().Size()
			if cw == mw && ch == mh {
				ebiten.SetWindowSize(800, 600)
				ebiten.SetWindowPosition(100, 100)
			} else {
				ebiten.SetWindowSize(mw, mh)
				ebiten.SetWindowPosition(0, 0)
			}
		}
		if g.minHov {
			ebiten.MinimizeWindow()
		}
	}

	ctrlS := (ebiten.IsKeyPressed(ebiten.KeyControlLeft) || ebiten.IsKeyPressed(ebiten.KeyControlRight)) && ebiten.IsKeyPressed(ebiten.KeyS)
	if ctrlS && !g.prevCtrlS {
		g.saveImage()
	}
	g.prevCtrlS = ctrlS

	g.prevLeft = left
	g.prevRight = right
	g.prevMid = mid
	return nil
}

func (g *Game) ptInSel(px, py float64) bool {
	if !g.hasSel {
		return false
	}
	minX := min(g.selStartX, g.selEndX)
	maxX := max(g.selStartX, g.selEndX)
	minY := min(g.selStartY, g.selEndY)
	maxY := max(g.selStartY, g.selEndY)
	return px >= minX && px <= maxX && py >= minY && py <= maxY
}

func (g *Game) screenToImg(sx, sy float64) (float64, float64) {
	iw := float64(g.img.Bounds().Dx())
	ih := float64(g.img.Bounds().Dy())
	ww, wh := ebiten.WindowSize()
	dw := iw * g.scale
	dh := ih * g.scale
	ix := (float64(ww) - dw) / 2 + g.offsetX
	iy := (float64(wh) - dh) / 2 + g.offsetY
	return (sx - ix) / g.scale, (sy - iy) / g.scale
}

func (g *Game) selImgCoords() (int, int, int, int) {
	x1, y1 := g.screenToImg(g.selStartX, g.selStartY)
	x2, y2 := g.screenToImg(g.selEndX, g.selEndY)
	iw := g.img.Bounds().Dx()
	ih := g.img.Bounds().Dy()
	minX := int(min(x1, x2))
	minY := int(min(y1, y2))
	maxX := int(max(x1, x2))
	maxY := int(max(y1, y2))
	if minX < 0 {
		minX = 0
	}
	if minY < 0 {
		minY = 0
	}
	if maxX > iw {
		maxX = iw
	}
	if maxY > ih {
		maxY = ih
	}
	return minX, minY, maxX, maxY
}

func (g *Game) crop() {
	if !g.hasSel || g.img == nil {
		return
	}
	minX, minY, maxX, maxY := g.selImgCoords()
	if maxX <= minX || maxY <= minY {
		return
	}
	sub := g.img.SubImage(image.Rect(minX, minY, maxX, maxY))
	g.img = ebiten.NewImageFromImage(sub)
	g.scale = 1.0
	g.offsetX, g.offsetY = 0, 0
	g.hasSel = false
	g.menuOpen = false
}

func (g *Game) copy() {
	if !g.hasSel || g.img == nil {
		return
	}
	minX, minY, maxX, maxY := g.selImgCoords()
	if maxX <= minX || maxY <= minY {
		return
	}
	sub := g.img.SubImage(image.Rect(minX, minY, maxX, maxY))
	b := sub.Bounds()
	rgba := image.NewRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			rgba.Set(x, y, sub.At(x, y))
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, rgba); err != nil {
		log.Println("copy:", err)
		return
	}
	clipboard.Write(clipboard.FmtImage, buf.Bytes())
	log.Println("copied to clipboard")
}

func (g *Game) saveImage() {
	if g.img == nil {
		return
	}
	fn := fmt.Sprintf("image_%d.png", time.Now().Unix())
	f, err := os.Create(fn)
	if err != nil {
		log.Println("save:", err)
		return
	}
	defer f.Close()
	if err := png.Encode(f, g.img); err != nil {
		log.Println("save:", err)
		return
	}
	log.Println("saved", fn)
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.NRGBA{R: 20, G: 20, B: 20, A: 255})

	if !g.hasImg {
		text.Draw(screen, "Drag image here or run: imageviewer <file>", basicfont.Face7x13, 20, 30, color.White)
		return
	}

	iw := float64(g.img.Bounds().Dx())
	ih := float64(g.img.Bounds().Dy())
	ww, wh := ebiten.WindowSize()

	dw := iw * g.scale
	dh := ih * g.scale
	x := (float64(ww) - dw) / 2 + g.offsetX
	y := (float64(wh) - dh) / 2 + g.offsetY

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(g.scale, g.scale)
	op.GeoM.Translate(x, y)
	screen.DrawImage(g.img, op)

	if g.hasSel || g.selecting {
		minX := min(g.selStartX, g.selEndX)
		maxX := max(g.selStartX, g.selEndX)
		minY := min(g.selStartY, g.selEndY)
		maxY := max(g.selStartY, g.selEndY)
		vector.DrawFilledRect(screen, float32(minX), float32(minY), float32(maxX-minX), float32(maxY-minY), color.NRGBA{R: 0, G: 120, B: 255, A: 40}, true)
		vector.StrokeRect(screen, float32(minX), float32(minY), float32(maxX-minX), float32(maxY-minY), 2, color.NRGBA{R: 0, G: 150, B: 255, A: 200}, true)
	}

	if g.hoverCtrl {
		vector.DrawFilledRect(screen, float32(ww-108), 10, 100, 30, color.NRGBA{R: 0, G: 0, B: 0, A: 100}, true)

		minBg := color.NRGBA{R: 255, G: 193, B: 7, A: 255}
		if g.minHov {
			minBg = color.NRGBA{R: 255, G: 215, B: 100, A: 255}
		}
		vector.DrawFilledCircle(screen, float32(ww-88), 25, 10, minBg, true)
		vector.StrokeLine(screen, float32(ww-88)-4, 25, float32(ww-88)+4, 25, 1.5, color.NRGBA{R: 50, G: 40, B: 10, A: 255}, true)

		maxBg := color.NRGBA{R: 40, G: 167, B: 69, A: 255}
		if g.maxHov {
			maxBg = color.NRGBA{R: 80, G: 200, B: 100, A: 255}
		}
		vector.DrawFilledCircle(screen, float32(ww-58), 25, 10, maxBg, true)
		vector.StrokeRect(screen, float32(ww-58)-4, 25-4, 8, 8, 1.5, color.NRGBA{R: 10, G: 40, B: 20, A: 255}, true)

		closeBg := color.NRGBA{R: 220, G: 53, B: 69, A: 255}
		if g.closeHov {
			closeBg = color.NRGBA{R: 255, G: 80, B: 80, A: 255}
		}
		vector.DrawFilledCircle(screen, float32(ww-28), 25, 10, closeBg, true)
		vector.StrokeLine(screen, float32(ww-28)-4, 25-4, float32(ww-28)+4, 25+4, 1.5, color.NRGBA{R: 50, G: 10, B: 10, A: 255}, true)
		vector.StrokeLine(screen, float32(ww-28)+4, 25-4, float32(ww-28)-4, 25+4, 1.5, color.NRGBA{R: 50, G: 10, B: 10, A: 255}, true)
	}

	if g.menuOpen {
		vector.DrawFilledRect(screen, float32(g.menuX), float32(g.menuY), 140, 105, color.NRGBA{R: 45, G: 45, B: 45, A: 255}, true)
		vector.StrokeRect(screen, float32(g.menuX), float32(g.menuY), 140, 105, 1, color.NRGBA{R: 80, G: 80, B: 80, A: 255}, true)
		text.Draw(screen, "Crop to selection", basicfont.Face7x13, g.menuX+10, g.menuY+20, color.NRGBA{R: 220, G: 220, B: 220, A: 255})
		text.Draw(screen, "Copy to clipboard", basicfont.Face7x13, g.menuX+10, g.menuY+55, color.NRGBA{R: 220, G: 220, B: 220, A: 255})
		text.Draw(screen, "Save image", basicfont.Face7x13, g.menuX+10, g.menuY+90, color.NRGBA{R: 220, G: 220, B: 220, A: 255})
	}
}

func (g *Game) Layout(ow, oh int) (int, int) {
	return ow, oh
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func main() {
	go func() {
		_ = clipboard.Init()
	}()

	g := &Game{scale: 1.0}
	if len(os.Args) > 1 {
		g.loadImage(os.Args[1])
	}

	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowDecorated(false)
	ebiten.SetWindowFloating(true)
	ebiten.SetWindowTitle("Image Viewer")

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

func (g *Game) loadImage(path string) {
	f, err := os.Open(path)
	if err != nil {
		log.Println("open:", err)
		return
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		f.Seek(0, 0)
		img, err = bmp.Decode(f)
		if err != nil {
			log.Println("decode:", err)
			return
		}
	}

	g.img = ebiten.NewImageFromImage(img)
	g.hasImg = true
	g.scale = 1.0
	g.offsetX, g.offsetY = 0, 0
	g.hasSel = false
	g.menuOpen = false
}
