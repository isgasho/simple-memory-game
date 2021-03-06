package main

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"os"
	"time"

	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/text"
	"golang.org/x/image/font"
)

const (
	screenSize, padSize, minLevel, maxLevel, newLevelDelayTicks, lightTicks, darkTicks, gameOverFlashTicks, gameOverTitle, gameOverInstruction = 480, 220, 1, 20, 60, 40, 20, 30, "GAME OVER", "TOUCH/CLICK TO START A NEW GAME"
)

const (
	demo = iota
	play
	over
)

var (
	sequence                               []int
	pads                                   []pad
	lastPressedPad                         *pad
	level, mode, currentIndex, tickCounter int
	gameOverMessage                        string
	smallFont, mediumFont, largeFont       font.Face

	blueDark    = &color.NRGBA{0x00, 0x00, 0x33, 0xff}
	blueLight   = &color.NRGBA{0x00, 0x00, 0xff, 0xff}
	greenDark   = &color.NRGBA{0x00, 0x33, 0x00, 0xff}
	greenLight  = &color.NRGBA{0x00, 0xff, 0x00, 0xff}
	redDark     = &color.NRGBA{0x33, 0x00, 0x00, 0xff}
	redLight    = &color.NRGBA{0xff, 0x00, 0x00, 0xff}
	yellowDark  = &color.NRGBA{0x33, 0x33, 0x00, 0xff}
	yellowLight = &color.NRGBA{0xff, 0xff, 0x00, 0xff}
)

type pad struct {
	x, y              float64
	on                bool
	imageOff, imageOn *ebiten.Image
}

func (pad pad) image() *ebiten.Image {
	if pad.on {
		return pad.imageOn
	}
	return pad.imageOff
}

func init() {
	rand.Seed(time.Now().UnixNano())

	tt, err := truetype.Parse(fonts.ArcadeN_ttf)
	if err != nil {
		panic(err)
	}
	smallFont = truetype.NewFace(tt, &truetype.Options{
		Size:    14,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	mediumFont = truetype.NewFace(tt, &truetype.Options{
		Size:    20,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	largeFont = truetype.NewFace(tt, &truetype.Options{
		Size:    48,
		DPI:     72,
		Hinting: font.HintingFull,
	})

	blueDarkImage, _ := ebiten.NewImage(padSize, padSize, ebiten.FilterDefault)
	blueDarkImage.Fill(blueDark)
	blueLightImage, _ := ebiten.NewImage(padSize, padSize, ebiten.FilterDefault)
	blueLightImage.Fill(blueLight)
	greenDarkImage, _ := ebiten.NewImage(padSize, padSize, ebiten.FilterDefault)
	greenDarkImage.Fill(greenDark)
	greenLightImage, _ := ebiten.NewImage(padSize, padSize, ebiten.FilterDefault)
	greenLightImage.Fill(greenLight)
	redDarkImage, _ := ebiten.NewImage(padSize, padSize, ebiten.FilterDefault)
	redDarkImage.Fill(redDark)
	redLightImage, _ := ebiten.NewImage(padSize, padSize, ebiten.FilterDefault)
	redLightImage.Fill(redLight)
	yellowDarkImage, _ := ebiten.NewImage(padSize, padSize, ebiten.FilterDefault)
	yellowDarkImage.Fill(yellowDark)
	yellowLightImage, _ := ebiten.NewImage(padSize, padSize, ebiten.FilterDefault)
	yellowLightImage.Fill(yellowLight)

	pads = append(pads, pad{0, 0, false, blueDarkImage, blueLightImage})
	pads = append(pads, pad{screenSize - padSize, 0, false, greenDarkImage, greenLightImage})
	pads = append(pads, pad{0, screenSize - padSize, false, redDarkImage, redLightImage})
	pads = append(pads, pad{screenSize - padSize, screenSize - padSize, false, yellowDarkImage, yellowLightImage})

	newGame()
}

func update(screen *ebiten.Image) error {
	switch mode {
	case demo:
		tickCounter++
		if currentIndex < level {
			if tickCounter == 1 {
				allPadsOff()
				pads[sequence[currentIndex]].on = true
			}
			if tickCounter == 1+lightTicks {
				allPadsOff()
			}
			if tickCounter == 1+lightTicks+darkTicks {
				tickCounter = 0
				currentIndex++
			}
		} else {
			currentIndex = 0
			mode = play
		}
	case play:
		allPadsOff()
		triggerPad := -1
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) || len(ebiten.TouchIDs()) > 0 {
			var posX, posY int
			if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
				posX, posY = ebiten.CursorPosition()
			} else {
				posX, posY = ebiten.TouchPosition(0)
			}
			pos := image.Point{posX, posY}
			pad := getPadAtPos(&pos)
			if pad != nil {
				pad.on = true
				lastPressedPad = pad
			} else {
				triggerPad = releaseLastPressedPad()
			}
		} else {
			triggerPad = releaseLastPressedPad()
		}

		if triggerPad >= 0 {
			if sequence[currentIndex] == triggerPad {
				if currentIndex+1 < level {
					currentIndex++
				} else {
					nextLevel()
				}
			} else {
				gameOver(fmt.Sprintf("YOU REACHED LEVEL %d", level))
			}
		}
	case over:
		tickCounter++
		if tickCounter > gameOverFlashTicks*2 {
			tickCounter = 0
		}
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) || len(ebiten.TouchIDs()) > 0 {
			newGame()
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		os.Exit(0)
	}

	if ebiten.IsDrawingSkipped() {
		return nil
	}

	switch mode {
	case over:
		text.Draw(screen, gameOverTitle, largeFont, (screenSize-text.MeasureString(gameOverTitle, largeFont).X)/2, screenSize*0.25, redLight)
		text.Draw(screen, gameOverMessage, mediumFont, (screenSize-text.MeasureString(gameOverMessage, mediumFont).X)/2, screenSize*0.5, greenLight)
		if tickCounter > gameOverFlashTicks {
			text.Draw(screen, gameOverInstruction, smallFont, (screenSize-text.MeasureString(gameOverInstruction, smallFont).X)/2, screenSize*0.75, yellowLight)
		}
	default:
		for _, pad := range pads {
			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(pad.x, pad.y)
			screen.DrawImage(pad.image(), opts)
		}
	}

	return nil
}

func allPadsOff() {
	for index := range pads {
		pads[index].on = false
	}
}

func getPadAtPos(pos *image.Point) *pad {
	for index, pad := range pads {
		if pos.In(pad.image().Bounds().Add(image.Point{int(pad.x), int(pad.y)})) {
			return &pads[index]
		}
	}
	return nil
}

func releaseLastPressedPad() int {
	returnVal := -1
	if lastPressedPad != nil {
		for index, pad := range pads {
			if pad == *lastPressedPad {
				returnVal = index
			}
		}
		lastPressedPad = nil
	}
	return returnVal
}

func newGame() {
	sequence = make([]int, maxLevel)
	for index := range sequence {
		sequence[index] = rand.Intn(4)
	}

	level = minLevel - 1
	nextLevel()
}

func nextLevel() {
	if level == maxLevel {
		gameOver(fmt.Sprintf("YOU BEAT ALL %v LEVELS!", level))
	} else {
		level++
		currentIndex = 0
		tickCounter = -newLevelDelayTicks
		mode = demo
		fmt.Println(fmt.Sprintf("LEVEL %v", level))
	}
}

func gameOver(message string) {
	fmt.Println(message)
	tickCounter = 0
	mode = over
	gameOverMessage = message
}

func main() {
	if err := ebiten.Run(update, screenSize, screenSize, 1, "Simple Memory Game"); err != nil {
		panic(err)
	}
}
