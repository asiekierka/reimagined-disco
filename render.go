package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	_ "image/png"
	"io/ioutil"
//	"math"
	"os"
//	"runtime"

	"github.com/go-gl/gl/v2.1/gl"
)

const Z_OFFSET = 1 / 64
const VIEW_DISTANCE = 5
const EYE_HEIGHT = 1.7

type Render struct {
	textures map[string]Texture
	blockSheet uint32
	buffers map[Position]*VertexBuffer
	toRefresh chan VertexRefreshRequest
}

type VertexRefreshRequest struct {
	pos	Position
	vbo	*VertexBuffer
}

type VertexBuffer struct {
	world          World
	data           []float32
	count          int32
	vbo            uint32
	vboCount       int32
	vboInit        bool
	refreshReady   bool
}

var playerLocal *Player

func (v *VertexBuffer) Append(q Quad) {
	for i := 0; i < 4; i++ {
		v.data = append(v.data, q.v[i].coord[0], q.v[i].coord[1], q.v[i].coord[2],
			q.normal[0], q.normal[1], q.normal[2],
			q.v[i].color[0], q.v[i].color[1], q.v[i].color[2],
			q.v[i].texcoord[0], q.v[i].texcoord[1])
	}
	v.count += 4
}

func (v *VertexBuffer) AppendFancy(q *Quad, coordOffset Vec3, texCoordOffset Vec2, lightLevel float32) {
	for i := 0; i < 4; i++ {
		v.data = append(v.data, q.v[i].coord[0] + coordOffset[0], q.v[i].coord[1] + coordOffset[1], q.v[i].coord[2] + coordOffset[2],
			q.normal[0], q.normal[1], q.normal[2],
			q.v[i].color[0] * lightLevel, q.v[i].color[1] * lightLevel, q.v[i].color[2] * lightLevel,
			q.v[i].texcoord[0] + texCoordOffset[0], q.v[i].texcoord[1] + texCoordOffset[1])
	}
	v.count += 4
}

func (v *VertexBuffer) Deinit() {
	gl.DeleteBuffers(1, &v.vbo)
	v.data = nil
}

func (v *VertexBuffer) Reset() {
	v.data = nil
	v.count = 0
}

func (v *VertexBuffer) Draw() {
	if v.count > 0 {
		if !v.vboInit {
			gl.GenBuffers(1, &v.vbo)
			v.vboInit = true
		}
		gl.BindBuffer(gl.ARRAY_BUFFER, v.vbo)

		if (v.refreshReady) {
 			gl.BufferData(gl.ARRAY_BUFFER, 4*len(v.data), gl.Ptr(v.data), gl.STATIC_DRAW)
			v.vboCount = v.count
			v.data = nil
			v.refreshReady = false
		}

		gl.VertexPointer(3, gl.FLOAT, 44, gl.PtrOffset(0))
		gl.TexCoordPointer(2, gl.FLOAT, 44, gl.PtrOffset(9*4))
		gl.NormalPointer(gl.FLOAT, 44, gl.PtrOffset(3*4))
		gl.ColorPointer(3, gl.FLOAT, 44, gl.PtrOffset(6*4))
		gl.DrawArrays(gl.QUADS, 0, v.vboCount)
	}
}

func nearestPow2(n int) int {
	t := uint32(n - 1)
	t |= t >> 1
	t |= t >> 2
	t |= t >> 4
	t |= t >> 8
	t |= t >> 16
	return int(t + 1)
}

func chunkRefreshLoop(r *Render) {
	a := 0
	for a == 0 {
		vbo := <- r.toRefresh
		vbo.vbo.Reset()
		renderChunk(r, vbo.pos, vbo.vbo)
		vbo.vbo.refreshReady = true
	}
}

func (r *Render) Init(width int32, height int32, debugtextures bool) {
	if err := gl.Init(); err != nil {
		panic(err)
	}

	r.toRefresh = make(chan VertexRefreshRequest, 64)
	go chunkRefreshLoop(r)
	go chunkRefreshLoop(r)

	r.buffers = make(map[Position]*VertexBuffer, 1000)
	r.textures = make(map[string]Texture)
	r.initTextures(debugtextures)
	setupScene()
	r.Resize(width, height)
}

func (r *Render) initTextures(debugtextures bool) {
	// TODO: not assume that all textures are going to be 16x16
	files, _ := ioutil.ReadDir("./textures/")
	textures := make(map[string]image.Image, 256)
	for _, texfn := range files {
		if texfn.IsDir() {
			continue
		}
		imgFile, err := os.Open("./textures/" + texfn.Name())
		if err == nil {
			img, _, err := image.Decode(imgFile)
			if err == nil {
				textures[texfn.Name()] = img
			}
			imgFile.Close()
		}
	}

	countSide := 1
	for countSide*countSide < len(textures) {
		countSide *= 2
	}
	rgba := image.NewRGBA(image.Rectangle{image.ZP, image.Pt(countSide << 4, countSide << 4)})
	pos := 0

	fmt.Printf("Initialized texture of size %d x %d\n", countSide << 4, countSide << 4)

	gl.Enable(gl.TEXTURE_2D)
	gl.GenTextures(1, &r.blockSheet)

	for name, img := range textures {
		pX := (pos % countSide)
		pY := (pos / countSide)
		dp := image.Pt(pX << 4, pY << 4)
		pos++

		draw.Draw(rgba, image.Rectangle{dp, dp.Add(image.Pt(16, 16))}, img, image.ZP, draw.Src)

		fmt.Printf("Loaded texture %s @ %d, %d\n", name, pX, pY)
		r.textures[name] = Texture{
			binding: r.blockSheet,
			minU: float32(pX) / float32(countSide),
			maxU: float32(pX + 1) / float32(countSide),
			minV: float32(pY) / float32(countSide),
			maxV: float32(pY + 1) / float32(countSide),
		}
	}

	if debugtextures {
		tmpFile, _ := os.Create("./blockSheet.png")
		png.Encode(tmpFile, rgba) 
		tmpFile.Close()
	}

	gl.BindTexture(gl.TEXTURE_2D, r.blockSheet)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix))
}

func (r *Render) markForUpdate(p Position) {
	buf, exists := r.buffers[p]
	if exists {
		r.toRefresh <- VertexRefreshRequest{pos: p, vbo: buf}
	}
}

func (r *Render) OnRenderUpdate(x int, y int, z int) {
	cx := x >> 4
	cy := y >> 4
	cz := z >> 4
	r.markForUpdate(Position{cx, cy, cz})
	if x & 15 == 0 {
		r.markForUpdate(Position{cx - 1, cy, cz})
	} else if x & 15 == 15 {
		r.markForUpdate(Position{cx + 1, cy, cz})
	}
	if y & 15 == 0 {
		r.markForUpdate(Position{cx, cy - 1, cz})
	} else if y & 15 == 15 {
		r.markForUpdate(Position{cx, cy + 1, cz})
	}
	if z & 15 == 0 {
		r.markForUpdate(Position{cx, cy, cz - 1})
	} else if z & 15 == 15 {
		r.markForUpdate(Position{cx, cy, cz + 1})
	}
}

func (r *Render) Deinit() {
	gl.DeleteTextures(1, &r.blockSheet)
}

func intMax(a int, b int) int {
	if a > b { return a } else { return b }
}

func isUsefulChunk(player *Player, p Position) bool {
	dist := intMax(intMax(p.x * 16 - int(player.pos[0]), p.y * 16 - int(player.pos[1])), p.z * 16 - int(player.pos[2]))
	return dist <= VIEW_DISTANCE*16
}

func dynamicChunkRender(r *Render, w World, p Position, vbo *VertexBuffer) {
}

func (r *Render) drawBlockVBOs(player *Player, w World) {
	gl.Enable(gl.TEXTURE_2D)
	gl.BindTexture(gl.TEXTURE_2D, r.blockSheet)

	pcx := int(player.pos[0]) >> 4
	pcy := int(player.pos[1]) >> 4
	pcz := int(player.pos[2]) >> 4

	// remove unused VBOs
	for pos, buf := range r.buffers {
		if !isUsefulChunk(player, pos) {
			buf.Deinit()
			delete(r.buffers, pos)
		}
	}

	gl.EnableClientState(gl.VERTEX_ARRAY)
	gl.EnableClientState(gl.TEXTURE_COORD_ARRAY)
	gl.EnableClientState(gl.NORMAL_ARRAY)
	gl.EnableClientState(gl.COLOR_ARRAY)

	// add necessary VBOs and render
	for y := -VIEW_DISTANCE; y <= VIEW_DISTANCE; y++ {
		for z := -VIEW_DISTANCE; z <= VIEW_DISTANCE; z++ {
			for x := -VIEW_DISTANCE; x <= VIEW_DISTANCE; x++ {
				pos := Position{x: pcx + x, y: pcy + y, z: pcz + z}
				if !w.IsLoaded(pos.x << 4, pos.y << 4, pos.z << 4) {
					continue
				}

				if _, exists := r.buffers[pos]; !exists {
					v := VertexBuffer{world: w}
					r.buffers[pos] = &v
					r.toRefresh <- VertexRefreshRequest{pos: pos, vbo: r.buffers[pos]}
				} else {
					r.buffers[pos].Draw()
				}
			}
		}
	}

	gl.DisableClientState(gl.VERTEX_ARRAY)
	gl.DisableClientState(gl.TEXTURE_COORD_ARRAY)
	gl.DisableClientState(gl.NORMAL_ARRAY)
	gl.DisableClientState(gl.COLOR_ARRAY)
}

func (r *Render) drawBlockHighlight(pos Position) {
	xMin := float32(pos.x) - Z_OFFSET
	xMax := float32(pos.x + 1) + Z_OFFSET
	yMin := float32(pos.y) - Z_OFFSET
	yMax := float32(pos.y + 1) + Z_OFFSET
	zMin := float32(pos.z) - Z_OFFSET
	zMax := float32(pos.z + 1) + Z_OFFSET

	gl.Disable(gl.TEXTURE_2D)
	gl.Enable(gl.LINE_SMOOTH)
	gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
	gl.LineWidth(2)
	gl.Begin(gl.QUADS)
	gl.Vertex3f(xMin, yMin, zMin)
	gl.Vertex3f(xMax, yMin, zMin)
	gl.Vertex3f(xMax, yMin, zMax)
	gl.Vertex3f(xMin, yMin, zMax)
	gl.Vertex3f(xMin, yMax, zMin)
	gl.Vertex3f(xMax, yMax, zMin)
	gl.Vertex3f(xMax, yMax, zMax)
	gl.Vertex3f(xMin, yMax, zMax)

	gl.Vertex3f(xMin, yMin, zMin)
	gl.Vertex3f(xMin, yMax, zMin)
	gl.Vertex3f(xMin, yMax, zMax)
	gl.Vertex3f(xMin, yMin, zMax)
	gl.Vertex3f(xMax, yMin, zMin)
	gl.Vertex3f(xMax, yMax, zMin)
	gl.Vertex3f(xMax, yMax, zMax)
	gl.Vertex3f(xMax, yMin, zMax)

	gl.Vertex3f(xMin, yMin, zMin)
	gl.Vertex3f(xMin, yMax, zMin)
	gl.Vertex3f(xMax, yMax, zMin)
	gl.Vertex3f(xMax, yMin, zMin)
	gl.Vertex3f(xMin, yMin, zMax)
	gl.Vertex3f(xMin, yMax, zMax)
	gl.Vertex3f(xMax, yMax, zMax)
	gl.Vertex3f(xMax, yMin, zMax)
	gl.End()
	gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
}

func (r *Render) Render(player *Player, w World) {
	// --- INIT ---
	playerLocal = player

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
	gl.Color4f(1, 1, 1, 1)

	// --- BLOCK AREA ---
	gl.PushMatrix()
	gl.Rotatef(player.pitch/DEG_RAD, 1, 0, 0)
	gl.Rotatef(player.yaw/DEG_RAD, 0, 1, 0)
	gl.Translatef(-player.pos[0], -player.pos[1]-EYE_HEIGHT, -player.pos[2])
	r.drawBlockVBOs(player, w)

	// draw block wireframe
	if pos, exists := player.GetHoverCoords(w); exists {
		r.drawBlockHighlight(pos)
	}
	gl.PopMatrix()

	// --- CLEANUP ---
}

func (r *Render) Resize(width int32, height int32) {
	ratio := float64(height) / float64(width) * 0.01

	gl.Viewport(0, 0, width, height)
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	if ratio > 1.0 {
		gl.Frustum(-(0.01 / ratio), (0.01 / ratio), -1, 1, 0.01, 512)
	} else {
		gl.Frustum(-0.01, 0.01, -ratio, ratio, 0.01, 512)
	}
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
}

func setupScene() {
	gl.Enable(gl.ALPHA_TEST)
	gl.Enable(gl.DEPTH_TEST)
	gl.Disable(gl.LIGHTING)

	gl.ClearColor(0.4, 0.6, 0.8, 0)
	gl.ClearDepth(1)
	gl.DepthFunc(gl.LEQUAL)

	gl.Fogi(gl.FOG_MODE, gl.LINEAR)
	gl.Fogf(gl.FOG_START, (VIEW_DISTANCE * 16) - 32)
	gl.Fogf(gl.FOG_END, (VIEW_DISTANCE * 16) - 8)
	fogcol := []float32{0.4, 0.6, 0.8, 1}
	gl.Fogfv(gl.FOG_COLOR, &fogcol[0])
	gl.Enable(gl.FOG)

	ambient := []float32{1, 1, 1, 1}
	gl.LightModelfv(gl.LIGHT_MODEL_AMBIENT, &ambient[0])
}

func isSolidSide(w World, x int, y int, z int, side Direction) bool {
	b := w.GetBlock(x, y, z)
	if b != nil {
		return b.IsSideSolid(side)
	} else {
		return false
	}
}

func renderQuad(vbo *VertexBuffer, x int, y int, z int, q *Quad, d Direction) {
	lightLevelScaler := []float32{0.55, 1.0, 0.85, 0.85, 0.7, 0.7}
	vbo.AppendFancy(q, Vec3{float32(x), float32(y), float32(z)}, Vec2{}, lightLevelScaler[int(d)])
}

func renderModel(vbo *VertexBuffer, x int, y int, z int, m Model) {
	for i := 0; i < 6; i++ {
		xOff := 0
		yOff := 0
		zOff := 0
		if i == 0 { yOff = -1 }
		if i == 1 { yOff = 1 }
		if i == 2 { xOff = -1 }
		if i == 3 { xOff = 1 }
		if i == 4 { zOff = -1 }
		if i == 5 { zOff = 1 }
		if !isSolidSide(vbo.world, x+xOff, y+yOff, z+zOff, Direction(i ^ 1)) {
			for _, quad := range m.faceQuads[i] {
				renderQuad(vbo, x, y, z, &quad, Direction(i))
			}
		}
	}
	for _, quad := range m.quads {
		renderQuad(vbo, x, y, z, &quad, UNKNOWN)
	}
}

func renderChunk(r *Render, p Position, vbo *VertexBuffer) {
	for y := 0; y < 16; y++ {
		py := p.y << 4 + y
		for z := 0; z < 16; z++ {
			pz := p.z << 4 + z
			for x := 0; x < 16; x++ {
				px := p.x << 4 + x
				block := vbo.world.GetBlock(px, py, pz)
				if block != nil {
					renderModel(vbo, px, py, pz, block.GetModel(r))
				}
			}
		}
	}
}
