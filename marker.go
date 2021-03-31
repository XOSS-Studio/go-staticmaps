// Copyright 2016, 2017 Florian Pigorsch. All rights reserved.
//
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sm

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/flopp/go-coordsparser"
	"github.com/fogleman/gg"
	"github.com/golang/geo/s2"
)

// Marker represents a marker on the map
type Marker struct {
	MapObject
	Position   s2.LatLng
	Color      color.Color
	Size       float64
	Label      string
	LabelColor color.Color
	im         image.Image
	offsetX float64
	offsetY float64
}

// NewMarker creates a new Marker
func NewMarker(pos s2.LatLng, col color.Color, size float64) *Marker {
	m := new(Marker)
	m.Position = pos
	m.Color = col
	m.Size = size
	m.Label = ""
	if Luminance(m.Color) >= 0.5 {
		m.LabelColor = color.RGBA{0x00, 0x00, 0x00, 0xff}
	} else {
		m.LabelColor = color.RGBA{0xff, 0xff, 0xff, 0xff}
	}
	return m
}

func NewImageMarker(pos s2.LatLng, im image.Image, offsetX, offsetY float64) *Marker {
	m := new(Marker)
	m.Position = pos
	m.im = im
	m.offsetX = offsetX
	m.offsetY = offsetY
	return m
}

func parseSizeString(s string) (float64, error) {
	switch {
	case s == "mid":
		return 16.0, nil
	case s == "small":
		return 12.0, nil
	case s == "tiny":
		return 8.0, nil
	}

	if floatValue, err := strconv.ParseFloat(s, 64); err == nil && floatValue > 0 {
		return floatValue, nil
	}

	return 0.0, fmt.Errorf("cannot parse size string: '%s'", s)
}

// ParseMarkerString parses a string and returns an array of markers
func ParseMarkerString(s string) ([]*Marker, error) {
	markers := make([]*Marker, 0)

	var markerColor color.Color = color.RGBA{0xff, 0, 0, 0xff}
	size := 16.0
	label := ""
	var labelColor color.Color

	for _, ss := range strings.Split(s, "|") {
		if ok, suffix := hasPrefix(ss, "color:"); ok {
			var err error
			markerColor, err = ParseColorString(suffix)
			if err != nil {
				return nil, err
			}
		} else if ok, suffix := hasPrefix(ss, "label:"); ok {
			label = suffix
		} else if ok, suffix := hasPrefix(ss, "size:"); ok {
			var err error
			size, err = parseSizeString(suffix)
			if err != nil {
				return nil, err
			}
		} else if ok, suffix := hasPrefix(ss, "labelcolor:"); ok {
			var err error
			labelColor, err = ParseColorString(suffix)
			if err != nil {
				return nil, err
			}
		} else {
			lat, lng, err := coordsparser.Parse(ss)
			if err != nil {
				return nil, err
			}
			m := NewMarker(s2.LatLngFromDegrees(lat, lng), markerColor, size)
			m.Label = label
			if labelColor != nil {
				m.SetLabelColor(labelColor)
			}
			markers = append(markers, m)
		}
	}
	return markers, nil
}

// SetLabelColor sets the color of the marker's text label
func (m *Marker) SetLabelColor(col color.Color) {
	m.LabelColor = col
}

// ExtraMarginPixels return the marker's left, top, right, bottom pixel extent.
func (m *Marker) ExtraMarginPixels() (float64, float64, float64, float64) {
	return 0.5*m.Size + 1.0, 1.5*m.Size + 1.0, 0.5*m.Size + 1.0, 1.0
}

// Bounds returns single point rect containing the marker's geographical position.
func (m *Marker) Bounds() s2.Rect {
	r := s2.EmptyRect()
	r = r.AddPoint(m.Position)
	return r
}

// Draw draws the object in the given graphical context.
func (m *Marker) Draw(gc *gg.Context, trans *Transformer) {
	if !CanDisplay(m.Position) {
		log.Printf("Marker coordinates not displayable: %f/%f", m.Position.Lat.Degrees(), m.Position.Lng.Degrees())
		return
	}
	x, y := trans.LatLngToXY(m.Position)
	if m.im == nil {
		gc.ClearPath()
		gc.SetLineJoin(gg.LineJoinRound)
		gc.SetLineWidth(1.0)

		radius := 0.5 * m.Size

		gc.DrawArc(x, y-m.Size, radius, (90.0+60.0)*math.Pi/180.0, (360.0+90.0-60.0)*math.Pi/180.0)
		gc.LineTo(x, y)
		gc.ClosePath()
		gc.SetColor(m.Color)
		gc.FillPreserve()
		gc.SetRGB(0, 0, 0)
		gc.Stroke()

		if m.Label != "" {
			gc.SetColor(m.LabelColor)
			gc.DrawStringAnchored(m.Label, x, y-m.Size, 0.5, 0.5)
		}
	} else {
		gc.DrawImage(m.im, int(x-m.offsetX), int(y-m.offsetY))
	}

}
