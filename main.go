package main

import (
	"fmt"
	"github.com/KEINOS/go-noise"
	"github.com/veandco/go-sdl2/sdl"
	"math"
	"math/rand"
	"runtime"
)

var renderer *sdl.Renderer

const (
	WindowWidth   = 1000
	WindowHeight  = 800
	TileSize      = 6
	EmptyTile     = 0
	Tree          = 1
	BurnedTree    = 2
	Fire          = 3
	BurningStage1 = 4
	BurningStage2 = 5
	BurningStage3 = 6
)

var forestMap [][]int32
var windDirectionDegreesMin int32
var windDirectionDegreesMax int32
var windSpreadProbability float32 = 0.5

func setupSDL(loop func()) {
	runtime.LockOSThread()

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Forest fire simulation",
		sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, WindowWidth, WindowHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	defer renderer.Destroy()

	defer println("\nFinishing simulation...")

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				running = false
				break
			case *sdl.KeyboardEvent:
				keyEvent := event.(*sdl.KeyboardEvent)
				if keyEvent.Type == sdl.KEYDOWN {
					switch keyEvent.Keysym.Sym {
					case sdl.K_q:
						running = false
					case sdl.K_r:
						newForest()
					case sdl.K_t:
						x, y := getRandomTile()
						for forestMap[x][y] != Tree {
							x, y = getRandomTile()
						}
						forestMap[x][y] = Fire
					}
				}
			}
		}

		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()

		loop()

		renderer.Present()

		sdl.Delay(16)
	}
}

func setDrawColorHex(hex string) {
	r, g, b := toRGB(hex)
	renderer.SetDrawColor(r, g, b, 255)
}

func drawTile(x, y int32, tileType int32) {
	switch tileType {
	case EmptyTile:
		setDrawColorHex("#000000")
	case Tree:
		setDrawColorHex("#228B22")
	case BurnedTree:
		setDrawColorHex("#3a0a00")
	case Fire:
		setDrawColorHex("#f7b633")
	case BurningStage1:
		setDrawColorHex("#b57332")
	case BurningStage2:
		setDrawColorHex("#a04d23")
	case BurningStage3:
		setDrawColorHex("#7a2617")
	}

	rect := sdl.Rect{X: x * TileSize, Y: y * TileSize, W: TileSize, H: TileSize}
	renderer.FillRect(&rect)
}

//goland:noinspection GoUnusedFunction
func generateTreesNoiseWithRandom() {
	forestMap = make([][]int32, WindowWidth/TileSize)
	for x := int32(0); x < WindowWidth/TileSize; x++ {
		forestMap[x] = make([]int32, WindowHeight/TileSize)
	}

	for pass := 0; pass < 5; pass++ {
		n, err := noise.New(noise.OpenSimplex, int64(pass)+rand.Int63())
		if err != nil {
			panic(err)
		}
		for x := int32(0); x < WindowWidth/TileSize; x++ {
			for y := int32(0); y < WindowHeight/TileSize; y++ {
				noiseValue := (n.Eval64(float64(x)/10, float64(y)/10) + 1) / 2
				if rand.Intn(100) < 4 || noiseValue > 0.75 {
					forestMap[x][y] = Tree
				}
			}
		}
	}
}

//goland:noinspection GoUnusedFunction
func generateTreesRandom() {
	forestMap = make([][]int32, WindowWidth/TileSize)
	for x := int32(0); x < WindowWidth/TileSize; x++ {
		forestMap[x] = make([]int32, WindowHeight/TileSize)
	}

	for x := int32(0); x < WindowWidth/TileSize; x++ {
		for y := int32(0); y < WindowHeight/TileSize; y++ {
			if rand.Intn(100) < 30 {
				forestMap[x][y] = Tree
			}
		}
	}
}

func newForest() {
	generateTreesRandom()
	// generateTreesNoiseWithRandom()
	windDirectionDegreesMin = rand.Int31n(360)
	windDirectionDegreesMax = windDirectionDegreesMin + randomMinMax(60, 120)*int32(rand.Intn(2)*2-1)
}

func getRandomTile() (int32, int32) {
	x := rand.Int31n(WindowWidth / TileSize)
	y := rand.Int31n(WindowHeight / TileSize)
	return x, y
}

func simulateDirectSpread(x, y int32) {
	switch forestMap[x][y] {
	case Fire:
		if x > 0 && forestMap[x-1][y] == Tree {
			forestMap[x-1][y] = Fire
		}
		if x < WindowWidth/TileSize-1 && forestMap[x+1][y] == Tree {
			forestMap[x+1][y] = Fire
		}
		if y > 0 && forestMap[x][y-1] == Tree {
			forestMap[x][y-1] = Fire
		}
		if y < WindowHeight/TileSize-1 && forestMap[x][y+1] == Tree {
			forestMap[x][y+1] = Fire
		}

		forestMap[x][y] = BurningStage1
	case BurningStage1:
		forestMap[x][y] = BurningStage2
	case BurningStage2:
		forestMap[x][y] = BurningStage3
	case BurningStage3:
		forestMap[x][y] = BurnedTree
	}
}

func simulateWind(x, y int32) {
	if Fire <= forestMap[x][y] && forestMap[x][y] <= BurningStage3 {
		for i := 1; i <= 5; i++ {
			windDirectionDegrees := windDirectionDegreesMin + randomMinMax(0, int32(math.Abs(float64(windDirectionDegreesMax-windDirectionDegreesMin))))
			windDirectionRadians := toRadians(windDirectionDegrees)
			windX := int32(float64(TileSize) * math.Cos(windDirectionRadians))
			windY := int32(float64(TileSize) * math.Sin(windDirectionRadians))

			if x+windX >= 0 && x+windX < WindowWidth/TileSize && y+windY >= 0 && y+windY < WindowHeight/TileSize {
				if forestMap[x+windX][y+windY] == Tree {
					if rand.Float32() < windSpreadProbability {
						forestMap[x+windX][y+windY] = Fire
					}
				}
			}
		}
	}
}

func main() {
	println("Shortcuts:" +
		"\n- Press 'Q' to quit the simulation." +
		"\n- Press 'R' to regenerate the forest." +
		"\n- Press 'T' to strike a thunderbolt on a random tree.")
	newForest()
	mainLoop := func() {
		for x := int32(0); x < WindowWidth/TileSize; x++ {
			for y := int32(0); y < WindowHeight/TileSize; y++ {
				drawTile(x, y, forestMap[x][y])

				simulateDirectSpread(x, y)
				simulateWind(x, y)
			}
		}
	}
	setupSDL(mainLoop)
}

func randomMinMax(min, max int32) int32 {
	return min + rand.Int31n(max-min+1)
}

func toRadians(degrees int32) float64 {
	return float64(degrees) * (math.Pi / 180)
}

func toRGB(hex string) (uint8, uint8, uint8) {
	var r, g, b uint8
	_, err := fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	if err != nil {
		panic(err)
	}
	return r, g, b
}
