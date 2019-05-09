// Package file handles the processing of game asset files.
package file

import (
	"errors"
	"fmt"
	"image"
	// import png for its side-effects
	_ "image/png"
	"os"

	"github.com/faiface/pixel"
)

var (
	assetsDir        = "assets/img/"
	imageAssetsStore = make(map[ImageFile]*pixel.PictureData)
)

// LoadAllAssets loads all assets into their appropriate assets store ready to be consumed.
func LoadAllAssets() error {
	// load all image assets
	for fileName := range imageFiles {
		if err := LoadPicture(fileName); err != nil {
			return fmt.Errorf("failed to load image asset: %s", err)
		}
	}

	return nil
}

// LoadPicture loads an image file from disk and stores it in the picture assets store.
func LoadPicture(fileName ImageFile) error {
	file, err := os.Open(assetsDir + fileName.String())
	if err != nil {
		return err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}
	imageAssetsStore[fileName] = pixel.PictureDataFromImage(img)
	return nil
}

// CreateSprite take a pre-loaded picture from the assets store and produces a new sprite from it.
func CreateSprite(fileName ImageFile) (*pixel.Sprite, error) {
	pic, ok := imageAssetsStore[fileName]
	if !ok {
		return nil, errors.New("image \"" + fileName.String() + "\" was not found in the assets store")
	}
	return pixel.NewSprite(pic, pic.Bounds()), nil
}

// ImageFile represents an image file name.
type ImageFile string

// String satisfies the Stringer interface to convert a file name into a plain string.
func (i ImageFile) String() string {
	return string(i)
}

// Image file name constants.
const (
	Player   ImageFile = "player.png"
	Grass    ImageFile = "grass.png"
	Sand     ImageFile = "sand.png"
	Water    ImageFile = "water.png"
	RoadNESW ImageFile = "road_nesw.png"
	RoadNES  ImageFile = "road_nes.png"
	RoadESW  ImageFile = "road_esw.png"
	RoadNSW  ImageFile = "road_nsw.png"
	RoadNEW  ImageFile = "road_new.png"
	RoadNE   ImageFile = "road_ne.png"
	RoadES   ImageFile = "road_es.png"
	RoadSW   ImageFile = "road_sw.png"
	RoadNW   ImageFile = "road_nw.png"
	RoadNS   ImageFile = "road_ns.png"
	RoadEW   ImageFile = "road_ew.png"
)

var imageFiles = map[ImageFile]bool{
	Player:   true,
	Grass:    true,
	Sand:     true,
	Water:    true,
	RoadNESW: true,
	RoadNES:  true,
	RoadESW:  true,
	RoadNSW:  true,
	RoadNEW:  true,
	RoadNE:   true,
	RoadES:   true,
	RoadSW:   true,
	RoadNW:   true,
	RoadNS:   true,
	RoadEW:   true,
}
