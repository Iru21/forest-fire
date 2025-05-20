package main

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"math"
	"math/rand"
	"runtime"
	"sync"
)

const (
	WindowWidth  = 1000
	WindowHeight = 800
	ChunkCount   = 16

	TileSize      = 4
	EmptyTile     = 0
	Tree          = 1
	BurnedTree    = 2
	Fire          = 3
	BurningStage1 = 4
	BurningStage2 = 5
	BurningStage3 = 6
)

var renderer *sdl.Renderer
var forestMap [][]int32
var nextMap [][]int32
var windDirectionDegreesMin int32
var windDirectionDegreesMax int32

var treeProbability float32 = 0.3
var windSpreadProbability float32 = 0.5
var limitThunderToCenter = true
var fullCircleWind = true

type Color struct {
	r uint8
	g uint8
	b uint8
}

var tileColors = map[int32]Color{
	EmptyTile:     toRGB("#000000"),
	Tree:          toRGB("#228B22"),
	BurnedTree:    toRGB("#3a0a00"),
	Fire:          toRGB("#000000"),
	BurningStage1: toRGB("#f7b633"),
	BurningStage2: toRGB("#b57332"),
	BurningStage3: toRGB("#a04d23"),
}

var tileTextures map[int32]*sdl.Texture

func createTileTextures() {
	tileTextures = make(map[int32]*sdl.Texture)
	for t, color := range tileColors {
		surf, _ := sdl.CreateRGBSurface(0, TileSize, TileSize, 32, 0, 0, 0, 0)
		surf.FillRect(nil, sdl.MapRGB(surf.Format, color.r, color.g, color.b))
		tex, _ := renderer.CreateTextureFromSurface(surf)
		tileTextures[t] = tex
		surf.Free()
	}
}

func drawTile(x, y int32, tileType int32) {
	tex := tileTextures[tileType]
	rect := sdl.Rect{X: x * TileSize, Y: y * TileSize, W: TileSize, H: TileSize}
	renderer.Copy(tex, nil, &rect)
}

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

	createTileTextures()

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
						strikeThunder()
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

func strikeThunder() {
	x, y := getRandomTile()
	for forestMap[x][y] != Tree ||
		(limitThunderToCenter && (!inRange(x, WindowWidth/TileSize/2-50, WindowWidth/TileSize/2+50) ||
			!inRange(y, WindowHeight/TileSize/2-50, WindowHeight/TileSize/2+50))) {
		x, y = getRandomTile()
	}
	forestMap[x][y] = Fire
}

func generateTreesRandom() {
	forestMap = make([][]int32, WindowWidth/TileSize)
	for x := int32(0); x < WindowWidth/TileSize; x++ {
		forestMap[x] = make([]int32, WindowHeight/TileSize)
	}

	maxTrees := WindowWidth * WindowHeight / (TileSize * TileSize)
	requiredTrees := int32(float32(maxTrees) * treeProbability)

	for i := int32(0); i < requiredTrees; i++ {
		x, y := getRandomTile()
		for forestMap[x][y] != EmptyTile {
			x, y = getRandomTile()
		}
		forestMap[x][y] = Tree
	}
}

func newForest() {
	generateTreesRandom()

	if fullCircleWind {
		windDirectionDegreesMin = 0
		windDirectionDegreesMax = 360
	} else {
		windDirectionDegreesMin = rand.Int31n(360)
		windDirectionDegreesMax = windDirectionDegreesMin + randomMinMax(60, 120)*int32(rand.Intn(2)*2-1)
	}
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
			nextMap[x-1][y] = Fire
		}
		if x < WindowWidth/TileSize-1 && forestMap[x+1][y] == Tree {
			nextMap[x+1][y] = Fire
		}
		if y > 0 && forestMap[x][y-1] == Tree {
			nextMap[x][y-1] = Fire
		}
		if y < WindowHeight/TileSize-1 && forestMap[x][y+1] == Tree {
			nextMap[x][y+1] = Fire
		}

		nextMap[x][y] = BurningStage1
	case BurningStage1:
		nextMap[x][y] = BurningStage2
	case BurningStage2:
		nextMap[x][y] = BurningStage3
	case BurningStage3:
		nextMap[x][y] = BurnedTree
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
						nextMap[x+windX][y+windY] = Fire
					}
				}
			}
		}
	}
}

func main() {
	println("Keybinds:" +
		"\n- Press 'Q' to quit the simulation." +
		"\n- Press 'R' to regenerate the forest." +
		"\n- Press 'T' to strike a thunderbolt on a random tree.")

	newForest()
	mainLoop := func() {
		var wg sync.WaitGroup
		chunkHeight := WindowHeight / TileSize / ChunkCount

		nextMap = make([][]int32, len(forestMap))
		for x := range forestMap {
			nextMap[x] = make([]int32, len(forestMap[x]))
			copy(nextMap[x], forestMap[x])
		}

		for i := 0; i < ChunkCount; i++ {
			startY := int32(i * chunkHeight)
			endY := startY + int32(chunkHeight)
			if i == ChunkCount-1 {
				endY = WindowHeight / TileSize
			}
			wg.Add(1)
			go func(startY, endY int32) {
				defer wg.Done()
				for x := int32(0); x < WindowWidth/TileSize; x++ {
					for y := startY; y < endY; y++ {
						simulateDirectSpread(x, y)
						simulateWind(x, y)
					}
				}
			}(startY, endY)
		}
		wg.Wait()
		forestMap, nextMap = nextMap, forestMap
		for x := int32(0); x < WindowWidth/TileSize; x++ {
			for y := int32(0); y < WindowHeight/TileSize; y++ {
				drawTile(x, y, forestMap[x][y])
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

func toRGB(hex string) Color {
	var r, g, b uint8
	_, err := fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	if err != nil {
		panic(err)
	}
	return Color{r: r, g: g, b: b}
}

func inRange(x, min, max int32) bool {
	return x >= min && x <= max
}
