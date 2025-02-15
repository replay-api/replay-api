package cdn

import "context"

type ImageCDNRepositoryMock struct {
}

func NewImageCDNRepositoryMock() ImageCDNRepositoryMock {
	return ImageCDNRepositoryMock{}
}

func (m ImageCDNRepositoryMock) Create(ctx context.Context, image []byte, name string, extension string) (string, error) {
	url := "https://rawusercontent.leetgaming.pro/media/" + name + extension

	return url, nil
}
