// Package world produces and maintains the game world.
package world

import (
	"fmt"
	"image/color"
	"math/rand"
	"sort"
	"sync"

	"github.com/aquilax/go-perlin"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"

	"github.com/jemgunay/procedural-game/file"
)

// Tile represents a single tile sprite and its corresponding properties.
type Tile struct {
	fileName     file.ImageFile
	sprite       *pixel.Sprite
	colourMask   color.Color
	visible      bool
	NoiseVal     float64
	roadMetaData string

	// the grid co-ordinate representation of the tile position
	gridPos pixel.Vec
	// the actual absolute tile position in pixels
	absPos pixel.Matrix
}

// SetSprite changes the tile's sprite to the specified image.
func (t *Tile) SetSprite(imageFile file.ImageFile) (err error) {
	sprite, err := file.CreateSprite(imageFile)
	if err != nil {
		return err
	}
	t.fileName = imageFile
	t.sprite = sprite
	return nil
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

// TileGrid is a concurrency safe map of tiles.
type TileGrid struct {
	tiles      map[string]*Tile
	terrainGen *perlin.Perlin
	randGen    *rand.Rand
	roadTiles  []*Tile
	sync.RWMutex
}

// NewTileGrid creates and initialises a new tile grid.
func NewTileGrid(seed int64) *TileGrid {
	rand.Seed(seed)
	return &TileGrid{
		tiles:      make(map[string]*Tile),
		terrainGen: perlin.NewPerlinRandSource(terrainPerlinAlpha, terrainPerlinBeta, terrainPerlinIterations, rand.NewSource(seed)),
		randGen:    rand.New(rand.NewSource(seed)),
	}
}

// createTile creates a new tile and inserts it into the tile grid given the tile image and x/y grid co-ordinates.
func (g *TileGrid) createTile(imageFile file.ImageFile, x, y int, z float64, mask color.Color) error {
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
		NoiseVal:   z,
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
func (g *TileGrid) Get(gridPos pixel.Vec) *Tile {
	g.RLock()
	tile := g.tiles[gridPos.String()]
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

			if err := g.createTile(tileImage, x, y, z, mask); err != nil {
				return fmt.Errorf("failed to create tile: %s", err)
			}
		}
	}
	fmt.Printf("Min Z: %v\nMax Z: %v\n", minZ, maxZ)

	// find points at the grass peaks (tiles where all neighbours have a smaller Z value)
	var peakTiles []*Tile
	for _, tile := range g.tiles {
		count := g.CheckNeighbours(tile, true, func(t1, t2 *Tile) bool {
			return t1.NoiseVal > t2.NoiseVal
		})
		if count == 8 {
			// check if new peak tile is too close to an existing peak tile
			tooClose := false
			for i := range peakTiles {
				if tile.gridPos.To(peakTiles[i].gridPos).Len() < 10 {
					tooClose = true
				}
			}

			if tooClose {
				continue
			}
			// tile is a peak tile
			//tile.colourMask = pixel.RGB(0, 0, 0)
			peakTiles = append(peakTiles, tile)
		}
	}

	// sort peak tiles by Z value
	sort.Slice(peakTiles, func(i2 int, j2 int) bool {
		return peakTiles[i2].NoiseVal > peakTiles[j2].NoiseVal
	})

	type distPair struct {
		tile        *Tile
		dist        float64
		connections int
	}
	for i := range peakTiles {
		// determine closest peak tiles for the current tile
		var dists []distPair

		for j := range peakTiles {
			// don't compare a tile against itself
			if i == j {
				continue
			}
			dist := peakTiles[i].gridPos.To(peakTiles[j].gridPos).Len()
			dists = append(dists, distPair{
				tile: peakTiles[j],
				dist: dist,
			})
		}

		// order all pairs by distance
		sort.Slice(dists, func(i2 int, j2 int) bool {
			return dists[i2].dist < dists[j2].dist
		})

		// cap n to max num of peak tiles
		neighbourCount := 2 + g.randGen.Intn(3)
		if neighbourCount > len(dists) {
			neighbourCount = len(dists)
		}
		// join closest n roads (n = 2 + ran(0, 3))
		for _, d := range dists[:neighbourCount] {
			g.joinTiles(peakTiles[i], d.tile)
		}
	}

	// set road sprite based on neighbouring road tiles
	for _, roadTile := range g.roadTiles {
		var (
			northTile    = g.Get(roadTile.gridPos.Add(pixel.V(0, 1)))
			eastTile     = g.Get(roadTile.gridPos.Add(pixel.V(1, 0)))
			southTile    = g.Get(roadTile.gridPos.Sub(pixel.V(0, 1)))
			westTile     = g.Get(roadTile.gridPos.Sub(pixel.V(1, 0)))
			roadTileName string
		)

		// compose road file name from neighbouring road tile compass positions
		if northTile != nil && northTile.roadMetaData == "ROAD" {
			roadTileName += "n"
		}
		if eastTile != nil && eastTile.roadMetaData == "ROAD" {
			roadTileName += "e"
		}
		if southTile != nil && southTile.roadMetaData == "ROAD" {
			roadTileName += "s"
		}
		if westTile != nil && westTile.roadMetaData == "ROAD" {
			roadTileName += "w"
		}

		// turn road tiles with only one road neighbour into straight roads
		switch roadTileName {
		case "n", "s":
			roadTileName = "ns"
		case "e", "w":
			roadTileName = "ew"
		}

		// change tile sprite to new road sprite
		if err := roadTile.SetSprite(file.ImageFile("road_" + roadTileName + ".png")); err != nil {
			return fmt.Errorf("failed to create road tile: %s", err)
		}
	}

	return nil
}

func (g *TileGrid) joinTiles(t1, t2 *Tile) {
	horizontalFirst := g.randGen.Intn(2) == 0
	t1X, t1Y := int(t1.gridPos.X), int(t1.gridPos.Y)
	t2X, t2Y := int(t2.gridPos.X), int(t2.gridPos.Y)

	if horizontalFirst {
		// draw horizontal roads
		if t1X > t2X {
			for i := t2X; i <= t1X; i++ {
				g.setTile(i, t1Y)
			}
		} else {
			for i := t1X; i <= t2X; i++ {
				g.setTile(i, t1Y)
			}
		}

		// draw vertical roads
		if t1Y > t2Y {
			for i := t2Y; i <= t1Y; i++ {
				g.setTile(t2X, i)
			}
		} else {
			for i := t1Y; i <= t2Y; i++ {
				g.setTile(t2X, i)
			}
		}

		return
	}

	// draw vertical roads
	if t1Y > t2Y {
		for i := t2Y; i <= t1Y; i++ {
			g.setTile(t1X, i)
		}
	} else {
		for i := t1Y; i <= t2Y; i++ {
			g.setTile(t1X, i)
		}
	}

	// draw horizontal roads
	if t1X > t2X {
		for i := t2X; i <= t1X; i++ {
			g.setTile(i, t2Y)
		}
	} else {
		for i := t1X; i <= t2X; i++ {
			g.setTile(i, t2Y)
		}
	}
}

func (g *TileGrid) setTile(X, Y int) {
	tile := g.Get(pixel.V(float64(X), float64(Y)))
	tile.roadMetaData = "ROAD"
	g.roadTiles = append(g.roadTiles, tile)
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
