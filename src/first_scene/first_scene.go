package first_scene

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
	"strconv"
	"time"
)

// Number of snowflakes
const SnowflakesCount = 100

var (
	// Flag of scene end
	IsOver = false

	// Binary file
	File *os.File  = nil

	// Channel for communication with music thread
	PlaySound  = make(chan bool, 2)

	// Sound bytes from binary file
	Sound = make([]byte, 0)

	// Vector of snowflakes
	Snowflakes = make([]*Snowflake, 0, SnowflakesCount)

	// Scene salt
	FirstSceneSalt = [settings.SaltLength]uint8{'f','i','r','s','t','s','c','e','n','e','.','.'}
)

// Function creating scene
// Append scene info to binary file
func CreateScene() error {
	// Open binary file with appending
	file, err := os.OpenFile(settings.BinaryPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Append salt to file
	err = binary.Write(file, binary.LittleEndian, &FirstSceneSalt)
	if err != nil {
		return err
	}

	// Append sound length to file
	soundLength := uint32(len(Music))
	err = binary.Write(file, binary.LittleEndian, soundLength)
	if err != nil {
		return err
	}

	// Append sound bytes to file
	err = binary.Write(file, binary.LittleEndian, Music)
	if err != nil {
		return err
	}

	// Append generated snowflakes to file
	for i := 0; i < SnowflakesCount; i++ {
		s := CreateSnowflake()

		err = binary.Write(file, binary.LittleEndian, s)
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

		// Try find first scene salt
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

			if salt == FirstSceneSalt {
				fmt.Println("[SUCCESS] First scene salt was found.")
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

	// Draw seconds on scene
	seconds := int(55 + 5 / 100.0 * float64(len(Snowflakes)))
	secondsStr := ""
	if seconds < 60 {
		secondsStr = "23:59:" + strconv.Itoa(seconds)
	} else {
		secondsStr = "00:00:00"
	}
	ctx.SetColor(colornames.White)
	ctx.DrawString(
		secondsStr,
		settings.CanvasWidth  / 2 - 20,
		settings.CanvasHeight / 2 + 10,
	)

	// Draw all snowflakes on canvas
	for i, s := range Snowflakes {
		if s != nil {
			s.Update()
			s.Show(ctx)

			if s.Y > settings.CanvasHeight {
				Snowflakes[i] = nil
			}
		}
	}

	// With 23% chance read next snowflake from binary file
	if rand.Float64() < 0.23 && len(Snowflakes) < SnowflakesCount {
		sf, readError := ReadSnowflakeFromBinary(File)
		if readError != nil {
			File.Close()
			return readError
		}

		Snowflakes = append(Snowflakes, sf)
	}

	// Checking end of scene (all snowflakes is down)
	IsOver = len(Snowflakes) >= SnowflakesCount
	for _, s := range Snowflakes {
		IsOver = IsOver && s == nil
	}

	// If scene is over - stop playing music and close file
	if IsOver {
		PlaySound <- false
		File.Close()
	}

	return nil
}
