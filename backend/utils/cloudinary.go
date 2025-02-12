package utils

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryService struct {
	cloudinary *cloudinary.Cloudinary
}

// Initialize a CloudinaryService instance
func NewCloudinaryService() (
	*CloudinaryService, error) {
	cloudName := os.Getenv("CLOUD_NAME")
	apiKey := os.Getenv("CLOUD_API_KEY")
	apiSecret := os.Getenv("CLOUD_API_SECRET")

	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Cloudinary: %w", err)
	}

	return &CloudinaryService{cloudinary: cld}, nil
}

// Upload an image to Cloudinary
func (cs *CloudinaryService) UploadImage(ctx context.Context,
	filePath string) (string, error) {
	uploadParams := uploader.UploadParams{
		Folder: "insta", // define a specific folder in Cloudinary
	}

	resp, err := cs.cloudinary.Upload.Upload(ctx, filePath, uploadParams)
	if err != nil {
		return "", fmt.Errorf("failed to upload image to Cloudinary: %w", err)
	}

	return resp.SecureURL, nil
}

// Delete an image from Cloudinary
func (cs *CloudinaryService) DeleteImage(ctx context.Context,
	publicID string) error {
	_, err := cs.cloudinary.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete image from Cloudinary: %w", err)
	}

	return nil
}
