// Package file manages the processing of files.
package file

import (
	"errors"
	"fmt"
	"image"
	_ "image/png"
	"io/ioutil"
	"os"

	"github.com/faiface/pixel"
)

var (
	assetsDir        = "../../assets/img/"
	imageAssetsStore = make(map[string]*pixel.PictureData)
)

// LoadAllAssets loads all assets into their appropriate assets store ready to be consumed.
func LoadAllAssets() error {
	files, err := ioutil.ReadDir(assetsDir)
	if err != nil {
		return fmt.Errorf("faield to read assets directory: %s", err)
	}

	// load all image assets
	for _, f := range files {
		if err = LoadPicture(f.Name()); err != nil {
			return fmt.Errorf("failed to load image asset: %s", err)
		}
	}
	return nil
}

// LoadPicture loads an image file from disk and stores it in the picture assets store.
func LoadPicture(fileName string) error {
	file, err := os.Open(assetsDir + fileName)
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

// ImageToSprite take a pre-loaded picture from the assets store and produces a new sprite from it.
func ImageToSprite(fileName string) (*pixel.Sprite, error) {
	pic, ok := imageAssetsStore[fileName]
	if !ok {
		return nil, errors.New("image \"" + fileName + "\" was not found in the assets store")
	}
	return pixel.NewSprite(pic, pic.Bounds()), nil
}
