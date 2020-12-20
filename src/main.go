package main

import (
	"drawing/first_scene"
	"drawing/second_scene"
	"drawing/settings"
	"fmt"
	"github.com/h8gi/canvas"
	"math/rand"
	"os"
	"time"
)

func main() {
	// Try remove last binary file
	removeError := os.Remove(settings.BinaryPath)
	if removeError != nil {
		fmt.Println("[WARNING] Removable binary file isn't exists.")
	} else {
		fmt.Println("[SUCCESS] Successfully remove a binary file.")
	}

	// Create new binary file
	createError := CreateCartoon()
	if createError != nil {
		fmt.Println("[ERROR] Couldn't write cartoon in binary file!")
		panic(createError)
	}

	// Create canvas for drawing
	c := canvas.NewCanvas(&canvas.CanvasConfig{
		Width:     settings.CanvasWidth,
		Height:    settings.CanvasHeight,
		FrameRate: settings.CanvasFPS,
		Title:     settings.CanvasTitle,
	})

	// Setup canvas for text output
	c.Setup(func(ctx *canvas.Context) {
		ctx.InvertY()
	})

	// Main draw function
	c.Draw(func(ctx *canvas.Context) {
		if !first_scene.IsOver {
			// Draw first scene until is over
			err := first_scene.DrawScene(ctx)

			// Handling not found error
			if err != nil {
				fmt.Println("[WARNING] Couldn't play first scene, salt not found!")
				first_scene.IsOver = true
			}
		} else if !second_scene.IsOver {
			// Draw second scene until is over
			err := second_scene.DrawScene(ctx)

			// Handling not found error
			if err != nil {
				fmt.Println("[WARNING] Couldn't play second scene, salt not found!")
				second_scene.IsOver = true
			}
		} else {
			// Exit if all scenes is over
			os.Exit(0)
		}
	})
}

// Function for creating scenes
// 91% - one scene => first scene chance = 91%
// 81% - two scenes => second scene chance = 81% / 91% = 89%
func CreateCartoon() error {
	// Set random seed
	rand.Seed(time.Now().UnixNano())

	// With 91% chance create first scene
	if rand.Float64() < settings.FirstSceneChance {
		// Try create first scene
		firstSceneError := first_scene.CreateScene()
		if firstSceneError != nil {
			return firstSceneError
		}

		// Echo info about creating
		fmt.Printf(
			"[SUCCESS] First scene successfully created with chance: %.2f\n",
			settings.FirstSceneChance,
		)

		// With 89% chance create second scene
		if rand.Float64() < settings.SecondSceneChange {
			// Try create second scene
			secondSceneError := second_scene.CreateScene()
			if secondSceneError != nil {
				return secondSceneError
			}

			// Echo info about creating
			fmt.Printf(
				"[SUCCESS] Second scene successfully created with chance: %.2f\n",
				settings.SecondSceneChange,
			)
		} else {
			// Echo info about bad second scene creation
			fmt.Println("[WARNING] Second scene was not created!")
		}
	} else {
		// Echo info about bad scenes creation
		fmt.Println("[WARNING] First scene was not created!")
		fmt.Println("[WARNING] Second scene was not created!")
	}

	return nil
}