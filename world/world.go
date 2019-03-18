// Package world produces and maintains the game world.
package world

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/aquilax/go-perlin"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"

	"github.com/jemgunay/game/file"
)

// Tile represents a single tile sprite and its corresponding properties.
type Tile struct {
	fileName file.ImageFile
	sprite   *pixel.Sprite
	visible  bool

	// the grid co-ordinate representation of the tile position
	gridPos pixel.Vec
	// the actual absolute tile position in pixels
	absPos pixel.Matrix
}

// TileGrid is a concurrency safe map of tiles.
type TileGrid struct {
	tiles map[string]*Tile
	sync.RWMutex
}

// NewTileGrid creates and initialises a new tile grid.
func NewTileGrid() *TileGrid {
	return &TileGrid{
		tiles: make(map[string]*Tile),
	}
}

// CreateTile creates a new tile and inserts it into the tile grid given the tile image and x/y grid co-ordinates.
func (g *TileGrid) CreateTile(imageFile file.ImageFile, x, y int) error {
	// create sprite
	sprite, err := file.CreateSprite(imageFile)
	if err != nil {
		return err
	}

	// create new tile
	newTile := &Tile{
		fileName: imageFile,
		sprite:   sprite,
		visible:  true,
		gridPos:  pixel.V(float64(x), float64(y)),
		absPos:   pixel.IM.Moved(pixel.V(float64(x*tileSize), float64(y*tileSize))),
	}

	// insert tile into tile grid
	g.Lock()
	g.tiles[newTile.gridPos.String()] = newTile
	g.Unlock()
	return nil
}

// Get retrieves a tile from a tile grid given the tile's grid position. Returns nil if tile does not exist for the
// provided grid position key.
func (g *TileGrid) Get(pos pixel.Vec) *Tile {
	g.RLock()
	tile := g.tiles[pos.String()]
	g.RUnlock()
	return tile
}

// Draw draws all of the tiles in the tile grid.
func (g *TileGrid) Draw(win *pixelgl.Window) {
	for _, tile := range g.tiles {
		if !tile.visible {
			continue
		}
		tile.sprite.Draw(win, tile.absPos)
	}
}

const (
	tileSize  = 100
	worldSize = 50

	// weight/noisiness
	perlinAlpha = 2.0
	// harmonic scaling/spacing
	perlinBeta = 2.0
	// number of iterations
	perlinIterations = 3
	// source of perlin noise randomness
	perlinSeed = int64(100)
)

// GenerateChunk generates a chunk of tiles.
func (g *TileGrid) GenerateChunk() error {
	var minZ, maxZ float64

	// generate perlin noise map
	p := perlin.NewPerlinRandSource(perlinAlpha, perlinBeta, perlinIterations, rand.NewSource(perlinSeed))
	for x := 0; x < worldSize; x++ {
		for y := 0; y < worldSize; y++ {
			z := p.Noise2D(float64(x)/10, float64(y)/10)

			if z < minZ {
				minZ = z
			}
			if z > maxZ {
				maxZ = z
			}

			if z < -0.45 {
				if err := g.CreateTile(file.Water, x, y); err != nil {
					return fmt.Errorf("failed to create tile: %s", err)
				}
			} else if z >= -0.45 && z < -0.1 {
				if err := g.CreateTile(file.Grass, x, y); err != nil {
					return fmt.Errorf("failed to create tile: %s", err)
				}
			} else {
				if err := g.CreateTile(file.RoadNESW, x, y); err != nil {
					return fmt.Errorf("failed to create tile: %s", err)
				}
			}
		}
	}

	// determine road tiles which
	for _, tile := range g.tiles {
		if tile.fileName != file.RoadNESW {
			continue
		}
		count := g.CheckNeighbours(tile, true, func(t1, t2 *Tile) bool {
			return t1.fileName == t2.fileName
		})
		if count == 8 {
			tile.visible = false
			//tile.visible = true
		}
	}

	fmt.Printf("Min Z: %v\nMax Z: %v\n", minZ, maxZ)
	return nil
}

// NeighbourFunc is used to compare two tiles based on the implemented criteria.
type NeighbourFunc func(t1, t2 *Tile) bool

// CheckNeighbours applies the specified NeighbourFunc to each of a tile's neighbours. If the NeighbourFunc evaluates to
// true for a given neighbouring tile, the returned count is incremented.
func (g *TileGrid) CheckNeighbours(tile *Tile, cornerNeighbours bool, checkFunc NeighbourFunc) (matchCount uint) {
	x, y := tile.gridPos.X, tile.gridPos.Y

	// north, east, south, west
	neighbours := [4]*Tile{
		g.Get(pixel.V(x, y+1)),
		g.Get(pixel.V(x+1, y)),
		g.Get(pixel.V(x, y-1)),
		g.Get(pixel.V(x-1, y)),
	}
	for _, n := range neighbours {
		if n != nil && checkFunc(tile, n) {
			matchCount++
		}
	}

	if !cornerNeighbours {
		return
	}

	// check cornering neighbours, i.e. north-east, south-east, south-west, north-west
	neighbours[0] = g.Get(pixel.V(x+1, y+1))
	neighbours[1] = g.Get(pixel.V(x+1, y-1))
	neighbours[2] = g.Get(pixel.V(x-1, y-1))
	neighbours[3] = g.Get(pixel.V(x-1, y+1))
	for _, n := range neighbours {
		if n != nil && checkFunc(tile, n) {
			matchCount++
		}
	}

	return
}
