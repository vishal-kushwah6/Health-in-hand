package cloudinary

import (
	storage "artGallery/internal/media/domain"
	"context"
	"io"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryStorage struct {
	cld *cloudinary.Cloudinary
}

func NewCloudinaryStorage(url string) (storage.Storage, error) {

	cld, err := cloudinary.NewFromURL(url)
	if err != nil {
		return nil, err
	}

	return &CloudinaryStorage{cld: cld}, nil
}

func (c *CloudinaryStorage) Upload(
	ctx context.Context,
	file io.Reader,
	folder string,
) (storage.UploadResult, error) {

	resp, err := c.cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder: folder,
	})

	if err != nil {
		return storage.UploadResult{}, err
	}

	return storage.UploadResult{
		URL: resp.SecureURL,
		Key: resp.PublicID,
	}, nil
}

func (c *CloudinaryStorage) Delete(
	ctx context.Context,
	key string,
) error {

	_, err := c.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: key,
	})

	return err
}
