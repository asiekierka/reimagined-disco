package main

import (
	"github.com/go-gl/gl/v2.1/gl"
)

type Texture struct {
	binding uint32
	minU float32
	maxU float32
	minV float32
	maxV float32
}

type Vertex struct {
	coord    Vec3
	texcoord Vec2
	color    Vec3
}

type Quad struct {
	v      [4]Vertex
	normal Vec3
}

type Model struct {
	faceQuads [6][]Quad
	quads []Quad
	dlist uint32
	hasDlist bool
}

func getQuadDirection(q *Quad) Direction {
	m := []int{2, 3, 0, 1, 4, 5}
	var d Direction
	var occ int
	for i := 0; i < 3; i++ {
		j := q.v[0].coord[i]
		if (j == 0 || j == 1) && j == q.v[1].coord[i] && q.v[1].coord[i] == q.v[2].coord[i] && q.v[2].coord[i] == q.v[3].coord[i] {
				occ++
				d = Direction(m[i * 2 + int(j)])
			}
	}
	if occ == 1 {
		return d
	} else {
		return UNKNOWN
	}
}

func (m *Model) freeDlist() {
	if m.hasDlist {
		gl.DeleteBuffers(1, &m.dlist)
		m.hasDlist = false
	}
}

func (m *Model) AddQuad(q Quad) {
	m.freeDlist()
	d := getQuadDirection(&q)
	if d == UNKNOWN {
		m.quads = append(m.quads, q)
	} else {
		m.faceQuads[int(d)] = append(m.faceQuads[int(d)], q)
	}
}

func (q *Quad) render() {
	gl.Normal3f(q.normal[0], q.normal[1], q.normal[2])
	for i := 0; i < 4; i++ {
		gl.Vertex3f(q.v[i].coord[0], q.v[i].coord[1], q.v[i].coord[2])
		gl.TexCoord2f(q.v[i].texcoord[0], q.v[i].texcoord[1])
		gl.Color3f(q.v[i].color[0], q.v[i].color[1], q.v[i].color[2])
	}
}

func (m *Model) Render() {
	if m.hasDlist {
		gl.CallList(m.dlist)
	} else {
		gl.NewList(m.dlist, gl.COMPILE_AND_EXECUTE)
		gl.Begin(gl.QUADS)
		for i := 0; i < 6; i++ {
			for _, quad := range m.faceQuads[i] {
				quad.render()
			}
		}
		for _, quad := range m.quads {
			quad.render()
		}
		gl.End()
		gl.EndList()
	}
}

func NewCubeModel(ta [6]Texture) Model {
	m := Model{}
	q := Quad{}
	c := Vec3{1, 1, 1}

	t := ta[5]
	q.normal = Vec3{0, 0, 1}
	q.v = [4]Vertex{
		Vertex{coord: Vec3{0, 1, 1}, texcoord: Vec2{t.minU, t.minV}, color: c},
		Vertex{coord: Vec3{1, 1, 1}, texcoord: Vec2{t.maxU, t.minV}, color: c},
		Vertex{coord: Vec3{1, 0, 1}, texcoord: Vec2{t.maxU, t.maxV}, color: c},
		Vertex{coord: Vec3{0, 0, 1}, texcoord: Vec2{t.minU, t.maxV}, color: c},
	}
	m.AddQuad(q)

	t = ta[4]
	q.normal = Vec3{0, 0, -1}
	q.v = [4]Vertex{
		Vertex{coord: Vec3{0, 1, 0}, texcoord: Vec2{t.minU, t.minV}, color: c},
		Vertex{coord: Vec3{1, 1, 0}, texcoord: Vec2{t.maxU, t.minV}, color: c},
		Vertex{coord: Vec3{1, 0, 0}, texcoord: Vec2{t.maxU, t.maxV}, color: c},
		Vertex{coord: Vec3{0, 0, 0}, texcoord: Vec2{t.minU, t.maxV}, color: c},
	}
	m.AddQuad(q)

	t = ta[1]
	q.normal = Vec3{0, 0, 1}
	q.v = [4]Vertex{
		Vertex{coord: Vec3{0, 1, 0}, texcoord: Vec2{t.minU, t.minV}, color: c},
		Vertex{coord: Vec3{1, 1, 0}, texcoord: Vec2{t.maxU, t.minV}, color: c},
		Vertex{coord: Vec3{1, 1, 1}, texcoord: Vec2{t.maxU, t.maxV}, color: c},
		Vertex{coord: Vec3{0, 1, 1}, texcoord: Vec2{t.minU, t.maxV}, color: c},
	}
	m.AddQuad(q)

	t = ta[0]
	q.normal = Vec3{0, 0, -1}
	q.v = [4]Vertex{
		Vertex{coord: Vec3{0, 0, 0}, texcoord: Vec2{t.minU, t.minV}, color: c},
		Vertex{coord: Vec3{1, 0, 0}, texcoord: Vec2{t.maxU, t.minV}, color: c},
		Vertex{coord: Vec3{1, 0, 1}, texcoord: Vec2{t.maxU, t.maxV}, color: c},
		Vertex{coord: Vec3{0, 0, 1}, texcoord: Vec2{t.minU, t.maxV}, color: c},
	}
	m.AddQuad(q)

	t = ta[3]
	q.normal = Vec3{1, 0, 0}
	q.v = [4]Vertex{
		Vertex{coord: Vec3{1, 1, 0}, texcoord: Vec2{t.minU, t.minV}, color: c},
		Vertex{coord: Vec3{1, 1, 1}, texcoord: Vec2{t.maxU, t.minV}, color: c},
		Vertex{coord: Vec3{1, 0, 1}, texcoord: Vec2{t.maxU, t.maxV}, color: c},
		Vertex{coord: Vec3{1, 0, 0}, texcoord: Vec2{t.minU, t.maxV}, color: c},
	}
	m.AddQuad(q)

	t = ta[2]
	q.normal = Vec3{-1, 0, 0}
	q.v = [4]Vertex{
		Vertex{coord: Vec3{0, 1, 0}, texcoord: Vec2{t.minU, t.minV}, color: c},
		Vertex{coord: Vec3{0, 1, 1}, texcoord: Vec2{t.maxU, t.minV}, color: c},
		Vertex{coord: Vec3{0, 0, 1}, texcoord: Vec2{t.maxU, t.maxV}, color: c},
		Vertex{coord: Vec3{0, 0, 0}, texcoord: Vec2{t.minU, t.maxV}, color: c},
	}
	m.AddQuad(q)

	return m
}
