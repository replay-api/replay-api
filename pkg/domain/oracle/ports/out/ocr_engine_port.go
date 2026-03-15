package oracle_out

import (
	"context"
)

// Region defines a rectangular area within an image for targeted OCR processing
type Region struct {
	X      int `json:"x" bson:"x"`           // Left offset in pixels
	Y      int `json:"y" bson:"y"`           // Top offset in pixels
	Width  int `json:"width" bson:"width"`   // Region width in pixels
	Height int `json:"height" bson:"height"` // Region height in pixels
}

// TextBlock represents a single text detection result from OCR
type TextBlock struct {
	Text        string  `json:"text"`
	Confidence  float64 `json:"confidence"`   // 0.0-1.0
	BoundingBox [4]int  `json:"bounding_box"` // [x, y, width, height]
}

// OCREnginePort defines the contract for extracting text from images
type OCREnginePort interface {
	// ExtractText runs OCR on the given image data, optionally cropping to a region
	ExtractText(ctx context.Context, imageData []byte, region *Region) ([]TextBlock, error)
}
