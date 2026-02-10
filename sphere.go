package main

import (
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Point struct {
	X, Y float32
}

type Line struct {
	points []Point
}

type Game struct {
	ballX, ballY   float32
	ballVX, ballVY float32
	ballRadius     float32
	firstBounce    bool

	bigCenterX, bigCenterY float32
	bigRadius              float32

	lines []*Line
}

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())

	bigCenterX := float32(400)
	bigCenterY := float32(300)
	bigRadius := float32(200)
	ballRadius := float32(10)

	startX := bigCenterX
	startY := bigCenterY

	return &Game{
		ballX:       startX,
		ballY:       startY,
		ballVX:      0,
		ballVY:      5,
		ballRadius:  ballRadius,
		firstBounce: true,

		bigCenterX: bigCenterX,
		bigCenterY: bigCenterY,
		bigRadius:  bigRadius,

		lines: []*Line{},
	}
}

func (g *Game) Update() error {
	g.ballX += g.ballVX
	g.ballY += g.ballVY

	dx := g.ballX - g.bigCenterX
	dy := g.ballY - g.bigCenterY
	distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))

	if distance >= g.bigRadius-g.ballRadius {
		nx := dx / distance
		ny := dy / distance

		impactX := g.bigCenterX + nx*g.bigRadius
		impactY := g.bigCenterY + ny*g.bigRadius

		g.ballX = g.bigCenterX + nx*(g.bigRadius-g.ballRadius)
		g.ballY = g.bigCenterY + ny*(g.bigRadius-g.ballRadius)

		if !g.firstBounce && len(g.lines) > 0 {
			currentLine := g.lines[len(g.lines)-1]
			currentLine.points = append(currentLine.points, Point{impactX, impactY})
		}

		if g.firstBounce {
			speed := float32(math.Sqrt(float64(g.ballVX*g.ballVX + g.ballVY*g.ballVY)))
			angle := rand.Float32() * 2 * math.Pi
			g.ballVX = speed * float32(math.Cos(float64(angle)))
			g.ballVY = speed * float32(math.Sin(float64(angle)))

			g.firstBounce = false
		} else {
			dot := g.ballVX*nx + g.ballVY*ny
			g.ballVX = g.ballVX - 2*dot*nx
			g.ballVY = g.ballVY - 2*dot*ny
		}

		newLine := &Line{
			points: []Point{{impactX, impactY}},
		}
		g.lines = append(g.lines, newLine)
	} else {
		if !g.firstBounce && len(g.lines) > 0 {
			currentLine := g.lines[len(g.lines)-1]
			currentLine.points = append(currentLine.points, Point{g.ballX, g.ballY})
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{245, 245, 245, 255})

	vector.StrokeCircle(screen, g.bigCenterX, g.bigCenterY, g.bigRadius, 2, color.RGBA{0, 0, 0, 255}, false)

	for _, l := range g.lines {
		for i := 0; i < len(l.points)-1; i++ {
			vector.StrokeLine(screen, l.points[i].X, l.points[i].Y,
				l.points[i+1].X, l.points[i+1].Y,
				2, color.RGBA{0, 0, 0, 255}, true)
		}
	}

	vector.FillCircle(screen, g.ballX, g.ballY, g.ballRadius, color.RGBA{0, 0, 0, 255}, true)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 800, 600
}

func main() {
	game := NewGame()
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("Echo Sphere")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
