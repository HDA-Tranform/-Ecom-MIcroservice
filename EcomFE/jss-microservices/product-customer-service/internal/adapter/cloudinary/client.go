package cloudinary

import (
	"context"
	"fmt"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/linhhuynhcoding/jss-microservices/product/config"
	"go.uber.org/zap"
)

type ICoundinaryClient interface {
	UploadImage(ctx context.Context, path string) (string, error)
}

type CloudinaryClient struct {
	cld    *cloudinary.Cloudinary
	logger *zap.Logger
	cfg    config.Config
}

func NewCloudinaryClient(logger *zap.Logger, cfg config.Config) ICoundinaryClient {
	cld, err := cloudinary.NewFromURL(cfg.CloudinaryConfig.ConnectString)
	if err != nil {
		logger.Error("failed to connect to cloudinary", zap.Error(err))
		return nil
	}
	// cld.Config.URL.Secure = true
	// cld.Config.URL = cfg.CloudinaryConfig.ConnectString
	// cld.Config.Cloud.APIKey = cfg.CloudinaryConfig.APIKey
	// cld.Config.Cloud.APISecret = cfg.CloudinaryConfig.APISecret
	// cld.Config.Cloud.CloudName = cfg.CloudinaryConfig.CloudName

	return &CloudinaryClient{
		cld:    cld,
		logger: logger,
		cfg:    cfg,
	}
}

func (c *CloudinaryClient) UploadImage(ctx context.Context, path string) (string, error) {
	c.logger.With(zap.String("func", "UploadImage"), zap.String("path", path))

	resp, err := c.cld.Upload.Upload(ctx, path, uploader.UploadParams{
		Folder:         "jss_folder",
		UniqueFilename: api.Bool(false),
		Overwrite:      api.Bool(true)})
	if err != nil {
		c.logger.Error("failed to upload image", zap.Error(err))
		return "", fmt.Errorf("failed to upload image: %v", err)
	}

	// TODO: return more information
	return resp.URL, nil
}
