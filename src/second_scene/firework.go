package second_scene

import (
	"encoding/binary"
	"github.com/h8gi/canvas"
	"image/color"
	"math"
	"math/rand"
	"os"
	"time"
)

// Number particles after boom!
const Particles = 120

type Firework struct {
	X            float64
	Y            float64
	Vy           float64
	Size         float64
	Acc          float64
	ExplodedTime float64
	Finished     bool
	Color        color.RGBA
}

// Function for generating next firework
func CreateFirework(width float64, height float64) *Firework {
	return &Firework{
		X:    rand.Float64() * width,
		Y:    height,
		Vy:   10 + rand.Float64()*5,
		Size: 4  + rand.Float64()*3,
		Acc:  -0.2,
		Color: color.RGBA{
			R: uint8(rand.Uint32()),
			G: uint8(rand.Uint32()),
			B: uint8(rand.Uint32()),
			A: 255,
		},
	}
}

// Function for getting next firework from binary file
func ReadFireworkFromBinary(file *os.File) (*Firework, error) {
	f := &Firework{}

	err := binary.Read(file, binary.LittleEndian, f)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// Function for updating one firework
func (f *Firework) Update() {
	f.Vy += f.Acc
	f.Y -= f.Vy

	if f.Vy <= 0 {
		f.ExplodedTime += 1
	}
}

// Function for drawing firework on canvas
func (f *Firework) Show(ctx *canvas.Context, height int) {
	if f.ExplodedTime == 0 {
		// Drawing firework
		ctx.SetColor(f.Color)
		ctx.DrawCircle(f.X, f.Y, f.Size)
		ctx.Fill()
	} else {
		// Finished flag
		finished := true

		// Drawing particles after boom!
		rand.Seed(0)
		for i := 0; i < Particles; i++ {
			radians := float64(i*360/Particles) * math.Pi / 180
			amplitude := f.ExplodedTime * 6

			cx := f.X + amplitude*math.Sin(radians)*rand.Float64()
			cy := f.Y + amplitude*math.Cos(radians)*rand.Float64()

			ctx.SetColor(f.Color)
			ctx.DrawCircle(cx, cy, f.Size/2)
			ctx.Fill()

			// Check if firework finished
			finished = finished && cy > float64(height)
		}
		rand.Seed(time.Now().UTC().UnixNano())

		// Set finished flag
		f.Finished = finished
	}
}
