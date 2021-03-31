package sm

import (
	"github.com/fogleman/gg"
	"github.com/golang/geo/s2"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"image"
	"image/color"
	"io/ioutil"
	"os"
	"testing"
)

func getImageFromFilePath(filePath string) (image.Image, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	image, _, err := image.Decode(f)
	return image, err
}

func TestMaker(t *testing.T) {
	fstart, err := getImageFromFilePath("./staticmap/start.png")
	if err != nil {
		panic(err)
	}
	fend, err := getImageFromFilePath("./staticmap/end.png")
	if err != nil {
		panic(err)
	}

	geoBytes, err := ioutil.ReadFile("a.geojson")
	if err != nil {
		panic(err)
	}
	fc, err := geojson.UnmarshalFeatureCollection(geoBytes)
	if err != nil {
		panic(err)
	}
	lineString := fc.Features[0].Geometry.(orb.LineString)
	startPoint := lineString[0]
	endpoint := lineString[len(lineString)-1]
	s2LineString := make([]s2.LatLng, len(lineString))
	for idx, item := range lineString {
		s2LineString[idx] = s2.LatLngFromDegrees(item[1], item[0])
	}

	ctx := NewContext()
	ctx.tileProvider = NewTileProviderXingzheInternal()
	ctx.SetSize(640*2, 280*2)

	ctx.AddObject(NewPath(s2LineString,
		color.RGBA{R: 17, G: 174, B: 250, A: 0xff}, 8.0),
	)
	offsetX, offsetY := 16.0, 16.0
	ctx.AddObject(
		NewImageMarker(
			s2.LatLngFromDegrees(startPoint[1], startPoint[0]),
			fstart,
			offsetX, offsetY),
	)
	ctx.AddObject(
		NewImageMarker(
			s2.LatLngFromDegrees(endpoint[1], endpoint[0]),
			fend,
			offsetX, offsetY),
	)

	img, err := ctx.Render()
	if err != nil {
		panic(err)
	}

	if err := gg.SavePNG("my-map.png", img); err != nil {
		panic(err)
	}
}
