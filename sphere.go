package main

import (
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Point struct {
	X, Y float32
}

type Ripple struct {
	X, Y   float32
	Radius float32
	Alpha  float32
	Hue    float32
}

type Game struct {
	ballX, ballY   float32
	ballVX, ballVY float32
	ballRadius     float32
	firstBounce    bool

	bigCenterX, bigCenterY float32
	bigRadius              float32
	wasColliding           bool

	impacts   []Point
	noteIndex int
	ticks     int
	ripples   []Ripple
}

func NewGame() *Game {
	bigCenterX := float32(400)
	bigCenterY := float32(300)
	bigRadius := float32(200)
	ballRadius := float32(8)

	return &Game{
		ballX:       bigCenterX,
		ballY:       bigCenterY,
		ballVX:      0,
		ballVY:      4,
		ballRadius:  ballRadius,
		firstBounce: true,

		bigCenterX:   bigCenterX,
		bigCenterY:   bigCenterY,
		bigRadius:    bigRadius,
		wasColliding: false,

		impacts: []Point{},
		ripples: []Ripple{},
	}
}

func hslToRGB(h, s, l float64) color.RGBA {
	c := (1 - math.Abs(2*l-1)) * s
	hp := h / 60.0
	x := c * (1 - math.Abs(math.Mod(hp, 2)-1))

	var r1, g1, b1 float64
	switch {
	case hp < 1:
		r1, g1, b1 = c, x, 0
	case hp < 2:
		r1, g1, b1 = x, c, 0
	case hp < 3:
		r1, g1, b1 = 0, c, x
	case hp < 4:
		r1, g1, b1 = 0, x, c
	case hp < 5:
		r1, g1, b1 = x, 0, c
	default:
		r1, g1, b1 = c, 0, x
	}

	m := l - c/2
	return color.RGBA{
		R: uint8((r1 + m) * 255),
		G: uint8((g1 + m) * 255),
		B: uint8((b1 + m) * 255),
		A: 255,
	}
}

func (g *Game) Update() error {
	g.ticks++

	g.ballX += g.ballVX
	g.ballY += g.ballVY

	dx := g.ballX - g.bigCenterX
	dy := g.ballY - g.bigCenterY
	dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))

	maxDist := g.bigRadius - g.ballRadius
	colliding := dist >= maxDist

	if colliding {
		nx := dx / dist
		ny := dy / dist

		if !g.wasColliding {
			impactX := g.bigCenterX + nx*g.bigRadius
			impactY := g.bigCenterY + ny*g.bigRadius

			g.impacts = append(g.impacts, Point{impactX, impactY})

			hue := float32(math.Mod(float64(g.noteIndex)*37, 360))
			g.ripples = append(g.ripples, Ripple{
				X: impactX, Y: impactY,
				Radius: 5, Alpha: 1.0, Hue: hue,
			})

			if g.firstBounce {
				speed := float32(math.Hypot(float64(g.ballVX), float64(g.ballVY)))
				g.ballVX = speed * 0.3
				g.ballVY = -speed * 0.95
				g.firstBounce = false
			} else {
				dot := g.ballVX*nx + g.ballVY*ny
				g.ballVX -= 2 * dot * nx
				g.ballVY -= 2 * dot * ny
			}

			dot := g.ballVX*nx + g.ballVY*ny
			if dot > 0 {
				g.ballVX -= 2 * dot * nx
				g.ballVY -= 2 * dot * ny
			}

			go PlayBounceNote(g.noteIndex)
			g.noteIndex++
		}

		g.ballX = g.bigCenterX + nx*(maxDist-1)
		g.ballY = g.bigCenterY + ny*(maxDist-1)
	}

	g.wasColliding = colliding

	alive := g.ripples[:0]
	for i := range g.ripples {
		g.ripples[i].Radius += 1.5
		g.ripples[i].Alpha -= 0.02
		if g.ripples[i].Alpha > 0 {
			alive = append(alive, g.ripples[i])
		}
	}
	g.ripples = alive

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{15, 15, 25, 255})

	pulseVal := 0.5 + 0.5*math.Sin(float64(g.ticks)*0.03)
	outerAlpha := uint8(80 + pulseVal*80)
	vector.StrokeCircle(screen, g.bigCenterX, g.bigCenterY, g.bigRadius, 2,
		color.RGBA{100, 140, 255, outerAlpha}, false)

	for i := 0; i < len(g.impacts)-1; i++ {
		hue := math.Mod(float64(i)*37, 360)
		c := hslToRGB(hue, 0.8, 0.55)
		vector.StrokeLine(screen, g.impacts[i].X, g.impacts[i].Y,
			g.impacts[i+1].X, g.impacts[i+1].Y,
			2, c, true)
	}

	if len(g.impacts) > 0 {
		last := g.impacts[len(g.impacts)-1]
		hue := math.Mod(float64(len(g.impacts)-1)*37, 360)
		c := hslToRGB(hue, 0.8, 0.55)
		vector.StrokeLine(screen, last.X, last.Y, g.ballX, g.ballY, 2, c, true)
	}

	for _, r := range g.ripples {
		a := uint8(r.Alpha * 200)
		rc := hslToRGB(float64(r.Hue), 0.9, 0.6)
		rc.A = a
		vector.StrokeCircle(screen, r.X, r.Y, r.Radius, 2, rc, true)
	}

	ballHue := math.Mod(float64(g.ticks)*1.5, 360)
	ballColor := hslToRGB(ballHue, 0.9, 0.6)

	glowColor := ballColor
	glowColor.A = 60
	vector.FillCircle(screen, g.ballX, g.ballY, g.ballRadius*2.5, glowColor, true)

	vector.FillCircle(screen, g.ballX, g.ballY, g.ballRadius, ballColor, true)

	vector.FillCircle(screen, g.ballX, g.ballY, g.ballRadius*0.4, color.RGBA{255, 255, 255, 180}, true)

	for i, p := range g.impacts {
		hue := math.Mod(float64(i)*37, 360)
		c := hslToRGB(hue, 0.9, 0.65)
		vector.FillCircle(screen, p.X, p.Y, 4, c, true)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 800, 600
}

func main() {
	Init()
	game := NewGame()
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("Echo Sphere")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
