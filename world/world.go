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
	sprite *pixel.Sprite
	// the co-ordinate representation of the tile position
	coordPos pixel.Vec
	// the actual absolute tile position in pixels
	absPos pixel.Matrix
}

// TileMap is a concurrency safe map of tiles.
type TileMap struct {
	tiles map[string]*Tile
	sync.RWMutex
}

// NewTileMap creates and initialises a new tile map.
func NewTileMap() *TileMap {
	return &TileMap{
		tiles: make(map[string]*Tile),
	}
}

// CreateTile creates a new tile and inserts it into the tile map.
func (m *TileMap) CreateTile(tileImageName string, x, y int) error {
	// create sprite
	sprite, err := file.CreateSprite(tileImageName + ".png")
	if err != nil {
		return err
	}

	// create new tile
	newTile := &Tile{
		sprite:   sprite,
		coordPos: pixel.V(float64(x), float64(y)),
		absPos:   pixel.IM.Moved(pixel.V(float64(x*tileSize), float64(y*tileSize))),
	}

	// insert tile into tile map
	m.Lock()
	m.tiles[newTile.absPos.String()] = newTile
	m.Unlock()
	return nil
}

// Get retrieves a tile from a tile map given the tile's position.
func (m *TileMap) Get(pos pixel.Vec) *Tile {
	m.RLock()
	tile := m.tiles[pos.String()]
	m.RUnlock()
	return tile
}

// Draw draws all of the tiles in the tile map.
func (m *TileMap) Draw(win *pixelgl.Window) {
	for _, tile := range m.tiles {
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
	seed             = int64(100)
)

// GenerateChunk generates a chunk of tiles.
func (m *TileMap) GenerateChunk() error {
	var minZ, maxZ float64

	// generate perlin noise map
	p := perlin.NewPerlinRandSource(perlinAlpha, perlinBeta, perlinIterations, rand.NewSource(seed))
	for x := 0; x < worldSize; x++ {
		for y := 0; y < worldSize; y++ {
			z := p.Noise2D(float64(x)/10, float64(y)/10)

			if z < minZ {
				minZ = z
			}
			if z > maxZ {
				maxZ = z
			}

			if z < 0 {
				if err := m.CreateTile("grass", x, y); err != nil {
					return fmt.Errorf("failed to create tile: %s", err)
				}
			} else {
				if err := m.CreateTile("road_nesw", x, y); err != nil {
					return fmt.Errorf("failed to create tile: %s", err)
				}
			}
		}
	}

	fmt.Printf("Min Z: %v\nMax Z: %v\n", minZ, maxZ)
	return nil
}
