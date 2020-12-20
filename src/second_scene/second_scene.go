package second_scene

import (
	"bytes"
	"drawing/settings"
	"encoding/binary"
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/h8gi/canvas"
	"golang.org/x/image/colornames"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"
)

// Number of all fireworks
const FireworksCount = 25

var (
	// Flag of scene end
	IsOver = false

	// Channel for communication with music thread
	PlaySound = make(chan bool, 2)

	// Sound bytes from binary file
	Sound = make([]byte, 0)

	// Binary file
	File *os.File = nil

	// Vector of fireworks
	Fireworks = make([]*Firework, 0, FireworksCount)

	// Scene salt
	SecondSceneSalt = [settings.SaltLength]uint8{'s','e','c','o','n','d','s','c','e','n','e','.'}
)

// Function creating scene
// Append scene info to binary file
func CreateScene() error {
	// Open binary file for appending info
	file, err := os.OpenFile(settings.BinaryPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Append salt to file
	err = binary.Write(file, binary.LittleEndian, &SecondSceneSalt)
	if err != nil {
		return err
	}

	// Append sound length info to file
	soundLength := uint32(len(Music))
	err = binary.Write(file, binary.LittleEndian, soundLength)
	if err != nil {
		return err
	}

	// Append sound to file
	err = binary.Write(file, binary.LittleEndian, Music)
	if err != nil {
		return err
	}

	// Append generated fireworks to file
	for i := 0; i < FireworksCount; i++ {
		f := CreateFirework(settings.CanvasWidth, settings.CanvasHeight)

		err = binary.Write(file, binary.LittleEndian, f)
		if err != nil {
			return err
		}
	}

	return nil
}

// Function for drawing scene on canvas
func DrawScene(ctx *canvas.Context) error {
	// If scene not initialized
	if File == nil {
		// Open created binary file
		file, openError := os.Open(settings.BinaryPath)
		if openError != nil {
			return openError
		}

		// Try find second scene salt
		var offset int64 = 0
		for {
			_, seekError := file.Seek(offset, 0)
			if seekError != nil {
				file.Close()
				return seekError
			}

			salt := [settings.SaltLength]uint8{}
			readError := binary.Read(file, binary.LittleEndian, &salt)
			if readError != nil {
				file.Close()
				return readError
			}

			if salt == SecondSceneSalt {
				fmt.Println("[SUCCESS] Second scene salt was found.")
				break
			}

			offset += 1
		}

		// Read sound length from binary file
		var soundLength uint32
		readSoundError := binary.Read(file, binary.LittleEndian, &soundLength)
		if readSoundError != nil {
			file.Close()
			return readSoundError
		}

		// Read sound bytes from binary file
		var temp byte
		for i := uint32(0); i < soundLength; i++ {
			readSoundError := binary.Read(file, binary.LittleEndian, &temp)
			if readSoundError != nil {
				file.Close()
				return readSoundError
			}

			Sound = append(Sound, temp)
		}

		// Start playing music in a separate thread
		// until music end or scene end
		go func() {
			// Create reader
			f := ioutil.NopCloser(bytes.NewReader(Sound))

			// Create stream
			streamer, format, err := mp3.Decode(f)
			if err != nil {
				log.Fatal(err)
			}
			defer streamer.Close()

			// Setup speaker
			playSoundError := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
			if playSoundError != nil {
				return
			}

			// Start playing
			speaker.Play(beep.Seq(streamer, beep.Callback(func() {
				PlaySound <- true
			})))

			// Wait and of playing or end of scene
			<- PlaySound

			speaker.Close()
		}()

		File = file
	}

	// Set bg color
	ctx.SetColor(colornames.Black)
	ctx.Clear()

	// Print text on canvas
	ctx.SetColor(colornames.White)
	ctx.DrawString(
		"NEW YEAR",
		float64(settings.CanvasWidth)/2.0-20,
		float64(settings.CanvasHeight)/2.0+5,
	)
	ctx.Fill()

	// Draw all fireworks on canvas
	for i, f := range Fireworks {
		if f != nil {
			f.Update()
			f.Show(ctx, settings.CanvasHeight)

			if f.Finished {
				Fireworks[i] = nil
			}
		}
	}

	// With 4% chance read next firework from binary file
	if rand.Float64() < 0.04 && len(Fireworks) < FireworksCount {
		f, readError := ReadFireworkFromBinary(File)
		if readError != nil {
			File.Close()
			return readError
		}

		Fireworks = append(Fireworks, f)
	}

	// Checking end of scene (all fireworks is down)
	IsOver = len(Fireworks) >= FireworksCount
	for _, f := range Fireworks {
		IsOver = IsOver && f == nil
	}

	// If scene is over - stop playing music and close file
	if IsOver {
		PlaySound <- false
		File.Close()
	}

	return nil
}
