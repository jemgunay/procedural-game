// Package file manages the processing of files.
package file

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
	Water    ImageFile = "water.png"
	RoadNESW ImageFile = "road_nesw.png"
	RoadNES  ImageFile = "road_nes.png"
	RoadESW  ImageFile = "road_esw.png"
	RoadNSW  ImageFile = "road_nsw.png"
	RoadNEW  ImageFile = "road_new.png"
	RoadNE   ImageFile = "road_ne.png"
	RoadSE   ImageFile = "road_se.png"
	RoadSW   ImageFile = "road_sw.png"
	RoadNW   ImageFile = "road_nw.png"
	RoadNS   ImageFile = "road_ns.png"
	RoadEW   ImageFile = "road_ew.png"
)

var imageFiles = map[ImageFile]bool{
	Player:   true,
	Grass:    true,
	Water:    true,
	RoadNESW: true,
	RoadNES:  true,
	RoadESW:  true,
	RoadNSW:  true,
	RoadNEW:  true,
	RoadNE:   true,
	RoadSE:   true,
	RoadSW:   true,
	RoadNW:   true,
	RoadNS:   true,
	RoadEW:   true,
}
