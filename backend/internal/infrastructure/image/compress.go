// internal/infrastructure/image/compress.go
package imageutil

import (
	"bytes"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"

	"github.com/disintegration/imaging"
)

type CompressOptions struct {
	MaxWidth  int
	MaxHeight int
	Quality   int // 1-100
}

func CompressAvatar(src io.Reader, opts CompressOptions) (io.Reader, error) {
	// 1. Decode image
	img, _, err := image.Decode(src)
	if err != nil {
		return nil, err
	}

	// 2. Resize — maintains aspect ratio, fits within bounds
	resized := imaging.Fit(img, opts.MaxWidth, opts.MaxHeight, imaging.Lanczos)

	// 3. Encode to JPEG with quality setting
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, resized, &jpeg.Options{Quality: opts.Quality})
	if err != nil {
		return nil, err
	}

	return &buf, nil
}
