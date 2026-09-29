//go:build ignore

// genicon draws the sweepkit icon. Run from the repository root:
//
//	go run scripts/genicon/main.go
//
// The same mark is drawn as SVG in frontend/dist/index.html.
package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
)

type rgb struct{ r, g, b float64 }

var (
	top    = rgb{0x14, 0xb8, 0xa6} // teal-500
	bottom = rgb{0x0f, 0x5f, 0x59}
	stroke = rgb{0xff, 0xff, 0xff}
	spark  = rgb{0xcc, 0xfb, 0xf1}
)

// Three sweep strokes, as capsules from (x0, y) to (x1, y), in a 1x1 box.
var strokes = [][3]float64{
	{0.22, 0.60, 0.40},
	{0.30, 0.78, 0.54},
	{0.22, 0.52, 0.68},
}

const (
	strokeRadius = 0.052
	margin       = 0.04
	corner       = 0.22
)

func main() {
	write(filepath.Join("build", "appicon.png"), 1024)
	write(filepath.Join("packaging", "linux", "sweepkit.png"), 256)
}

func write(path string, size int) {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	const ss = 4 // supersampling per axis
	for py := 0; py < size; py++ {
		for px := 0; px < size; px++ {
			var r, g, b, a float64
			for sy := 0; sy < ss; sy++ {
				for sx := 0; sx < ss; sx++ {
					x := (float64(px) + (float64(sx)+0.5)/ss) / float64(size)
					y := (float64(py) + (float64(sy)+0.5)/ss) / float64(size)
					c, cov := sample(x, y)
					r += c.r * cov
					g += c.g * cov
					b += c.b * cov
					a += cov
				}
			}
			n := float64(ss * ss)
			if a > 0 {
				img.SetNRGBA(px, py, color.NRGBA{uint8(r / a), uint8(g / a), uint8(b / a), uint8(255 * a / n)})
			}
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		log.Fatal(err)
	}
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
}

// sample returns the colour and coverage at a point.
func sample(x, y float64) (rgb, float64) {
	if roundedRect(x, y) > 0 {
		return rgb{}, 0
	}
	t := (x + y) / 2
	c := rgb{lerp(top.r, bottom.r, t), lerp(top.g, bottom.g, t), lerp(top.b, bottom.b, t)}
	for _, s := range strokes {
		if capsule(x, y, s[0], s[1], s[2]) <= strokeRadius {
			return stroke, 1
		}
	}
	if sparkle(x-0.74, y-0.30) {
		return spark, 1
	}
	return c, 1
}

func roundedRect(x, y float64) float64 {
	cx, cy := math.Abs(x-0.5), math.Abs(y-0.5)
	h := 0.5 - margin - corner
	dx, dy := math.Max(cx-h, 0), math.Max(cy-h, 0)
	return math.Hypot(dx, dy) - corner
}

// capsule is the distance from (x, y) to the horizontal segment (x0..x1, cy).
func capsule(x, y, x0, x1, cy float64) float64 {
	return math.Hypot(x-math.Max(x0, math.Min(x, x1)), y-cy)
}

// sparkle is a four-pointed star made of two thin diamonds.
func sparkle(x, y float64) bool {
	const long, thin = 0.11, 0.028
	diamond := func(a, b float64) bool { return math.Abs(x)/a+math.Abs(y)/b <= 1 }
	return diamond(long, thin) || diamond(thin, long)
}

func lerp(a, b, t float64) float64 { return a + (b-a)*t }
