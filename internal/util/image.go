package util

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	_ "image/png"
	"math"

	"golang.org/x/image/draw"
)

type ProcessedImage struct {
	Bytes    []byte
	MimeType string
}

func ProcessImageToJPEGMaxBytes(input []byte, maxBytes int) (*ProcessedImage, error) {
	if len(input) == 0 {
		return nil, errors.New("empty image")
	}

	img, _, err := image.Decode(bytes.NewReader(input))
	if err != nil {
		return nil, err
	}

	srcBounds := img.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()
	if srcW <= 0 || srcH <= 0 {
		return nil, errors.New("invalid image bounds")
	}

	qualitySteps := []int{85, 75, 65, 55, 45, 35}

	scale := 1.0
	for attempt := 0; attempt < 10; attempt++ {
		var out image.Image = img
		if scale < 0.999 {
			dstW := int(math.Max(1, math.Round(float64(srcW)*scale)))
			dstH := int(math.Max(1, math.Round(float64(srcH)*scale)))
			dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
			draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
			out = dst
		}

		for _, q := range qualitySteps {
			var buf bytes.Buffer
			if err := jpeg.Encode(&buf, out, &jpeg.Options{Quality: q}); err != nil {
				return nil, err
			}
			if buf.Len() <= maxBytes {
				return &ProcessedImage{Bytes: buf.Bytes(), MimeType: "image/jpeg"}, nil
			}
		}

		scale *= 0.85
		if scale < 0.1 {
			break
		}
	}

	return nil, errors.New("image too large after resize")
}
