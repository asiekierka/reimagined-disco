package main

import (
	"flag"
	"fmt"
	_ "image/png"
	"log"
	"math"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/barnex/fmath"
	"github.com/go-gl/glfw/v3.1/glfw"
)

var cpuprofile = flag.Bool("cpuprofile", false, "write cpu profile to file")
var heapprofile = flag.Bool("heapprofile", false, "write heap profile to file")
var debugtextures = flag.Bool("debugtextures", false, "write texture sheet to file")

type Player struct {
	pos Vec3
	gravity float32
	yaw   float32
	pitch float32
}

type Average struct {
	data []float64
	pos  int
}

func NewAverage(len int) Average {
	return Average{make([]float64, len), 0}
}

func (a *Average) Push(v float64) {
	a.data[a.pos] = v
	a.pos++
	if a.pos == len(a.data) {
		a.pos = 0
	}
}

func (a Average) Get() float64 {
	var t float64
	for i := 0; i < len(a.data); i++ {
		t += a.data[i]
	}
	return t / float64(len(a.data))
}

var (
	fps       Average
	player    Player
	movementX float32
	movementZ float32
	render    Render
	lastMx    float64
	lastMy    float64
)

const DEG_RAD = math.Pi / 180

func onMouse(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
	if action == glfw.Press {
		if button == glfw.MouseButtonLeft {
			breakBlock()
		} else if button == glfw.MouseButtonRight {
			placeBlock()
		}
	}
}

func onResize(w *glfw.Window, width int, height int) {
	render.Resize(int32(width), int32(height))
}

func onMove(w *glfw.Window, x float64, y float64) {
	player.yaw += float32((x - lastMx) / 1000)
	player.pitch += float32((y - lastMy) / 1000)
	if player.pitch < -(math.Pi / 2) {
		player.pitch = -(math.Pi / 2)
	} else if player.pitch > (math.Pi / 2) {
		player.pitch = (math.Pi / 2)
	}
	lastMx = x
	lastMy = y
}

func (player Player) GetHoverCoords(w World) (Position, bool) {
	stepX := -fmath.Sin(-player.yaw) * fmath.Cos(player.pitch) * 0.2
	stepY := -fmath.Sin(player.pitch) * 0.2
 	stepZ := -fmath.Cos(-player.yaw) * fmath.Cos(player.pitch) * 0.2
	bX := player.pos[0]
	bY := player.pos[1] + EYE_HEIGHT
	bZ := player.pos[2]
	for stepCount := 100; stepCount > 0; stepCount-- {
		if w.GetBlock(int(bX), int(bY), int(bZ)) != nil {
			return Position{int(bX), int(bY), int(bZ)}, true
		} else {
			bX += stepX
			bY += stepY
			bZ += stepZ
		}		
	}
	return Position{}, false
}

func breakBlock() {
	if pos, exists := player.GetHoverCoords(&w); exists {
		w.SetBlock(pos.x, pos.y, pos.z, nil)
	}
}

func placeBlock() {
	stepX := -fmath.Sin(-player.yaw) * fmath.Cos(player.pitch) * 0.2
	stepY := -fmath.Sin(player.pitch) * 0.2
 	stepZ := -fmath.Cos(-player.yaw) * fmath.Cos(player.pitch) * 0.2
	bX := player.pos[0]
	bY := player.pos[1] + EYE_HEIGHT
	bZ := player.pos[2]
	for stepCount := 100; stepCount > 0; stepCount-- {
		if w.GetBlock(int(bX), int(bY), int(bZ)) != nil {
			bX -= stepX
			bY -= stepY
			bZ -= stepZ
			w.SetBlock(int(bX), int(bY), int(bZ), br.ByName("gold_block"))
			return
		} else {
			bX += stepX
			bY += stepY
			bZ += stepZ
		}
	}
}

func onKey(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyW {
		if action == glfw.Release {
			movementX = 0.0
		} else {
			movementX = 0.12
		}
	}
	if key == glfw.KeyS {
		if action == glfw.Release {
			movementX = 0.0
		} else {
			movementX = -0.12
		}
	}
	if key == glfw.KeyA {
		if action == glfw.Release {
			movementZ = 0.0
		} else {
			movementZ = -0.12
		}
	}
	if key == glfw.KeyD {
		if action == glfw.Release {
			movementZ = 0.0
		} else {
			movementZ = 0.12
		}
	}
	if key == glfw.KeySpace {
		if action == glfw.Press {
			player.gravity = 0.3
		}
	}
}

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

var (
	w WorldFlat
	br BlockRegistry
)

func main() {
	flag.Parse()

	fps = NewAverage(256)
	player.pos = Vec3{8, MAP_H + 16, 8}

	br = NewBlockRegistry()
	br.Register(&BlockSimple{name: "grass", textures: [6]string{"dirt.png", "grass.png", "grass_side.png", "grass_side.png", "grass_side.png", "grass_side.png"}})
	br.Register(&BlockSimple{name: "dirt", textures: [6]string{"dirt.png", "dirt.png", "dirt.png", "dirt.png", "dirt.png", "dirt.png"}})
	br.Register(&BlockSimple{name: "stone", textures: [6]string{"stone.png", "stone.png", "stone.png", "stone.png", "stone.png", "stone.png"}})
	br.Register(&BlockSimple{name: "gold_block", textures: [6]string{"gold_block.png", "gold_block.png", "gold_block.png", "gold_block.png", "gold_block.png", "gold_block.png"}})

	w = NewWorldFlat(br)

	fmt.Printf("Loading...\n")
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	window, err := glfw.CreateWindow(800, 600, "GM-M1-142, strona 12", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()
	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	window.SetKeyCallback(onKey)
	window.SetCursorPosCallback(onMove)
	window.SetFramebufferSizeCallback(onResize)
	window.SetMouseButtonCallback(onMouse)

	//glfw.SwapInterval(0)

	render.Init(800, 600, *debugtextures)
	defer render.Deinit()

	w.RegisterRenderListener(&render)

	if *cpuprofile {
		fmt.Printf("CPU profiling ON!")
	        f, err := os.Create("cpu.prof")
	        if err != nil {
       			log.Fatal(err)
        	}
        	pprof.StartCPUProfile(f)
        	defer pprof.StopCPUProfile()
	}

	for !window.ShouldClose() {
		t := time.Now()
		render.Render(&player, &w)
		window.SwapBuffers()
		glfw.PollEvents()
		nanoTime := time.Since(t)
		movementLX := movementX * float32(nanoTime) / (16 * 1000000)
		movementLZ := movementZ * float32(nanoTime) / (16 * 1000000)
		gravityLD := 0.5 * float32(nanoTime) / (1000 * 1000000)
		fpsNow := float64(1000000000) / float64(nanoTime)
		fps.Push(fpsNow)
		//fmt.Printf("%.2f (%.2f) [%.2f %.2f %.2f]\n", fps.Get(), nanoTime, player.pos[0], player.pos[1], player.pos[2])

		// player.pos[1] - player.pitch * movementLD,
		// fmath.Cos(player.pitch)

		// fall
		if w.GetBlock(int(player.pos[0]), int(fmath.Ceil(player.pos[1])) - 1, int(player.pos[2])) == nil {
			player.gravity -= gravityLD
		} else if player.gravity < 0 {
			player.gravity = 0
		}

		if player.gravity < 0 {
			bY := player.pos[1]
			for w.GetBlock(int(player.pos[0]), int(fmath.Ceil(player.pos[1])), int(player.pos[2])) == nil && player.pos[1] > (bY + player.gravity) {
				player.pos[1] -= 0.0025
			}
		} else if player.gravity > 0 {
			bY := player.pos[1]
			for w.GetBlock(int(player.pos[0]), int(fmath.Ceil(player.pos[1])), int(player.pos[2])) == nil && player.pos[1] < (bY + player.gravity) {
				player.pos[1] += 0.0025
			}
		}

		nx := player.pos[0] - fmath.Sin(-player.yaw) * movementLX + fmath.Cos(-player.yaw) * movementLZ
		nz := player.pos[2] - fmath.Cos(-player.yaw) * movementLX - fmath.Sin(-player.yaw) * movementLZ

		if w.GetBlock(int(nx), int(fmath.Ceil(player.pos[1])), int(nz)) == nil {
			player.pos[0] = nx
			player.pos[2] = nz
		}
	}

	if *heapprofile {
		f, err := os.Create("heap.prof")
	        if err != nil {
	            log.Fatal(err)
	        }
		pprof.WriteHeapProfile(f)
	        f.Close()
	}
}
