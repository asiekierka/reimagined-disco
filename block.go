package main

type Block interface {
	New() Block
	Name() string
	GetModel(r *Render) Model
	GetBoundingBox() BoundingBox
	IsSideSolid(d Direction) bool
}

type BlockRegistry struct {
	idBlock []Block
	nameBlock map[string]Block
	nameId map[string]int
	maxId int
}

type BlockSimple struct {
	name string
	textures [6]string
	models map[*Render]Model
}

func NewBlockRegistry() BlockRegistry {
	return BlockRegistry{
		idBlock: make([]Block, 256, 256),
		nameId: make(map[string]int, 256),
		nameBlock: make(map[string]Block, 256),
		maxId: 1,
	}
}

func (b *BlockRegistry) ByName(s string) Block {
	return b.nameBlock[s]
}

func (b *BlockRegistry) ByID(i int) Block {
	return b.idBlock[i]
}

func (b *BlockRegistry) GetID(bb Block) int {
	return b.nameId[bb.Name()]
}

func (b *BlockRegistry) allocID() int {
	b.maxId++
	for b.idBlock[b.maxId] != nil {
		b.maxId++
	}
	return b.maxId - 1
}

func (b *BlockRegistry) Register(bb Block) {
	id := b.allocID()
	b.nameBlock[bb.Name()] = bb
	b.nameId[bb.Name()] = id
	b.idBlock[id] = bb
}

func (b *BlockSimple) Name() string {
	return b.name
}

func (b *BlockSimple) GetModel(r *Render) Model {
	if b.models == nil {
		b.models = make(map[*Render]Model, 1)
	}
	if v, ok := b.models[r]; ok {
		return v
	}
	b.models[r] = NewCubeModel([6]Texture{
		r.textures[b.textures[0]],
		r.textures[b.textures[1]],
		r.textures[b.textures[2]],
		r.textures[b.textures[3]],
		r.textures[b.textures[4]],
		r.textures[b.textures[5]],
	})
	return b.models[r]
}

func (b *BlockSimple) GetBoundingBox() BoundingBox {
	return BoundingBox{Vec3{0,0,0}, Vec3{1,1,1}}
}

func (b *BlockSimple) IsSideSolid(d Direction) bool {
	return true
}

func (b *BlockSimple) New() Block {
	return b
}
