// Package world produces and maintains the game world.
package world

import (
	"fmt"
	"image/color"
	"math/rand"
	"sync"

	"github.com/aquilax/go-perlin"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"

	"github.com/jemgunay/game/file"
)

// Tile represents a single tile sprite and its corresponding properties.
type Tile struct {
	fileName   file.ImageFile
	sprite     *pixel.Sprite
	colourMask color.Color
	visible    bool

	// the grid co-ordinate representation of the tile position
	gridPos pixel.Vec
	// the actual absolute tile position in pixels
	absPos pixel.Matrix
}

// TileGrid is a concurrency safe map of tiles.
type TileGrid struct {
	tiles      map[string]*Tile
	terrainGen *perlin.Perlin
	roadGen    *perlin.Perlin
	sync.RWMutex
}

const (
	// just grater than 200 to overlap, preventing stitching glitch
	tileSize  = 201
	chunkSize = 50

	// weight/noisiness
	terrainPerlinAlpha = 2.0
	// harmonic scaling/spacing
	terrainPerlinBeta = 1.0
	// number of iterations
	terrainPerlinIterations = 3
)

// NewTileGrid creates and initialises a new tile grid.
func NewTileGrid(seed int64) *TileGrid {
	return &TileGrid{
		tiles:      make(map[string]*Tile),
		terrainGen: perlin.NewPerlinRandSource(terrainPerlinAlpha, terrainPerlinBeta, terrainPerlinIterations, rand.NewSource(seed)),
		roadGen:    perlin.NewPerlinRandSource(1.0, terrainPerlinBeta, terrainPerlinIterations, rand.NewSource(seed)),
	}
}

// createTile creates a new tile and inserts it into the tile grid given the tile image and x/y grid co-ordinates.
func (g *TileGrid) createTile(imageFile file.ImageFile, x, y int, mask color.Color) error {
	// create sprite
	sprite, err := file.CreateSprite(imageFile)
	if err != nil {
		return err
	}

	// create new tile
	absPos := pixel.V(float64(x)*tileSize, float64(y)*tileSize)
	newTile := &Tile{
		fileName:   imageFile,
		sprite:     sprite,
		colourMask: mask,
		visible:    true,
		gridPos:    pixel.V(float64(x), float64(y)),
		absPos:     pixel.IM.Scaled(pixel.V(float64(x), float64(y)), 2.0).Moved(absPos),
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
		tile.sprite.DrawColorMask(win, tile.absPos, tile.colourMask)
	}
}

// GenerateChunk generates a chunk of tiles.
func (g *TileGrid) GenerateChunk() error {
	var minZ, maxZ = 10.0, -10.0

	// generate perlin noise map
	for x := 0; x < chunkSize; x++ {
		for y := 0; y < chunkSize; y++ {
			// add one to scale z between 0 and 2
			z := g.terrainGen.Noise2D(float64(x)/11, float64(y)/11) + 1

			if z < minZ {
				minZ = z
			}
			if z > maxZ {
				maxZ = z
			}

			var (
				tileImage file.ImageFile
				mask      color.Color
				waterMax  = 0.66
				sandMax   = 0.8
			)
			if z < waterMax {
				tileImage = file.Water
			} else if z >= waterMax && z < sandMax {
				tileImage = file.Sand
			} else if z >= sandMax {
				tileImage = file.Grass
				// shade grass tiles based on height
				grassMin, outMin, outMax := 1.1, 1.0, 0.9
				maskVal := (z-waterMax)*(outMax-outMin)/(grassMin-waterMax) + outMin
				mask = pixel.RGB(maskVal, maskVal, maskVal)
			}

			roadZ := g.roadGen.Noise2D(float64(x)/11, float64(y)/11) + 1
			if roadZ > 1.5 {
				tileImage = file.RoadNESW
			}

			if err := g.createTile(tileImage, x, y, mask); err != nil {
				return fmt.Errorf("failed to create tile: %s", err)
			}
		}
	}

	/*for _, tile := range g.tiles {
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
	}*/

	fmt.Printf("Min Z: %v\nMax Z: %v\n", minZ, maxZ)
	return nil
}

// NeighbourFunc is used to compare two tiles based on the implemented criteria.
type NeighbourFunc func(t1, t2 *Tile) bool

// CheckNeighbours applies the specified NeighbourFunc to each of a tile's neighbours. If the NeighbourFunc evaluates to
// true for a given neighbouring tile, the returned count is incremented.
func (g *TileGrid) CheckNeighbours(tile *Tile, cornerNeighbours bool, checkFunc NeighbourFunc) (matchCount uint) {
	x, y := tile.gridPos.XY()

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
