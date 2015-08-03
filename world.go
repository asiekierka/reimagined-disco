package main

import (
	"math/rand"
	"github.com/larspensjo/Go-simplex-noise/simplexnoise"
)

const (
	MAP_W = 512
	MAP_H = 128
	MAP_D = 512
)

type WorldFlat struct {
	blocks []int16
	blockReg BlockRegistry
	renderListeners []RenderListener
}

type World interface {
	IsValid(int, int, int) bool
	IsLoaded(int, int, int) bool
	GetBlock(int, int, int) Block
	SetBlock(int, int, int, Block)
}

type BlockAccess interface {
	GetBlock(int, int, int) Block
	SetBlock(int, int, int, Block)
}

type RenderListener interface {
	OnRenderUpdate(int, int, int)
}

func NewWorldFlat(blockReg BlockRegistry) WorldFlat {
	w := WorldFlat{
		blocks: make([]int16, MAP_W*MAP_H*MAP_D),
		blockReg: blockReg,
	}

	seed := float64(rand.Intn(20000000))
	arr := [6]float64{0.004, 0.008, 0.016, 0.032, 0.064, 0.0128}
	arr2 := [6]float64{32, 16, 8, 4, 2, 1}
	for z := 0; z < MAP_D; z++ {
		for x := 0; x < MAP_W; x++ {
			heightF := float64(MAP_H / 2)
			for i := 0; i < len(arr); i++ {
				heightF += simplexnoise.Noise3(float64(x) * arr[i], float64(z) * arr[i], seed + float64(i * 1000000)) * arr2[i]
			}
			height := int(heightF)
			t := blockReg.ByName("stone")
			for h := 0; h < MAP_H; h++ {
				if h == (height - 3) {
					t = blockReg.ByName("dirt")
				} else if h == height {
					t = blockReg.ByName("grass")
				} else if h > height {
					t = nil
				}
				if t != nil {
					w.SetBlock(x, h, z, t.New())
				}
			}
		}
	}

	return w
}

func pos(x int, y int, z int) int {
	return (y*MAP_D + z)*MAP_W + x
}

func (w *WorldFlat) IsLoaded(x int, y int, z int) bool {
	return w.IsValid(x, y, z)
}

func (w *WorldFlat) IsValid(x int, y int, z int) bool {
	return x >= 0 && y >= 0 && z >= 0 && x < MAP_W && y < MAP_H && z < MAP_D
}

func (w *WorldFlat) GetBlock(x int, y int, z int) Block {
	if w.IsValid(x,y,z) {
		return w.blockReg.ByID(int(w.blocks[pos(x,y,z)]))
	} else {
		return nil
	}
}

func (w *WorldFlat) RegisterRenderListener(r RenderListener) {
	if w.renderListeners == nil {
		w.renderListeners = make([]RenderListener, 1, 1)
	}
	w.renderListeners = append(w.renderListeners, r)
}

func (w *WorldFlat) SetBlock(x int, y int, z int, block Block) {
	if w.IsValid(x,y,z) {
		if block == nil {
			w.blocks[pos(x,y,z)] = 0
		} else {
			w.blocks[pos(x,y,z)] = int16(w.blockReg.GetID(block))
		}
		for _, listener := range w.renderListeners {
			if listener != nil {
				listener.OnRenderUpdate(x, y, z)
			}
		}
	}
}
