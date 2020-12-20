package first_scene

import (
	"drawing/settings"
	"encoding/binary"
	"github.com/h8gi/canvas"
	"math"
	"math/rand"
	"os"
)

type Snowflake struct {
	X        float64
	Y        float64
	Angle    float64
	Size     float64
	Radius   float64
	Lifetime float64
}

// Function for generating new snowflake
func CreateSnowflake() *Snowflake {
	return &Snowflake{
		X:      rand.Float64() * settings.CanvasWidth,
		Y:      0,
		Angle:  2 * math.Pi * rand.Float64(),
		Size:   1.5 + rand.Float64()*2,
		Radius: math.Sqrt(rand.Float64() * math.Pow(settings.CanvasWidth/2, 2)),
	}
}

// Function for updating one snowflake by lifetime
func (s *Snowflake) Update() {
	// Calculate angle
	angle := 0.2 * (s.Lifetime + 200) * s.Angle

	// Change position
	s.X = settings.CanvasWidth/2 + s.Radius*math.Sin(angle)
	s.Y += s.Size

	// Increase lifetime
	s.Lifetime += 0.02
}

// Function for drawing snowflake on canvas
func (s *Snowflake) Show(ctx *canvas.Context) {
	ctx.SetRGBA(1, 1, 1, s.Size / 3.5)
	ctx.DrawCircle(s.X, s.Y, s.Size)
	ctx.Fill()
}

// Function for getting next snowflake from binary file
func ReadSnowflakeFromBinary(file *os.File) (*Snowflake, error) {
	s := &Snowflake{}

	err := binary.Read(file, binary.LittleEndian, s)
	if err != nil {
		return nil, err
	}

	return s, nil
}