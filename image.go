package xai

import (
	"context"

	v1 "github.com/roelfdiedericks/xai-go/proto/xai/api/v1"
)

// ImageFormat specifies the output format for generated images.
type ImageFormat int

const (
	// ImageFormatURL returns images as URLs.
	ImageFormatURL ImageFormat = iota
	// ImageFormatBase64 returns images as base64-encoded strings.
	ImageFormatBase64
)

func (f ImageFormat) toProto() v1.ImageFormat {
	switch f {
	case ImageFormatBase64:
		return v1.ImageFormat_IMG_FORMAT_BASE64
	default:
		return v1.ImageFormat_IMG_FORMAT_URL
	}
}

// ImageAspectRatio specifies the aspect ratio for generated images.
type ImageAspectRatio int

const (
	// ImageAspectRatio1x1 generates a square image.
	ImageAspectRatio1x1 ImageAspectRatio = iota
	// ImageAspectRatio16x9 generates a wide image.
	ImageAspectRatio16x9
	// ImageAspectRatio9x16 generates a tall image.
	ImageAspectRatio9x16
	// ImageAspectRatio4x3 generates a classic landscape image.
	ImageAspectRatio4x3
	// ImageAspectRatio3x4 generates a classic portrait image.
	ImageAspectRatio3x4
)

func (ar ImageAspectRatio) toProto() v1.ImageAspectRatio {
	switch ar {
	case ImageAspectRatio16x9:
		return v1.ImageAspectRatio_IMG_ASPECT_RATIO_16_9
	case ImageAspectRatio9x16:
		return v1.ImageAspectRatio_IMG_ASPECT_RATIO_9_16
	case ImageAspectRatio4x3:
		return v1.ImageAspectRatio_IMG_ASPECT_RATIO_4_3
	case ImageAspectRatio3x4:
		return v1.ImageAspectRatio_IMG_ASPECT_RATIO_3_4
	default:
		return v1.ImageAspectRatio_IMG_ASPECT_RATIO_1_1
	}
}

// ImageResolution specifies the resolution for generated images.
type ImageResolution int

const (
	// ImageResolution1K generates ~1024px images.
	ImageResolution1K ImageResolution = iota
	// ImageResolution2K generates ~2048px images (upscaled).
	ImageResolution2K
)

func (r ImageResolution) toProto() v1.ImageResolution {
	switch r {
	case ImageResolution2K:
		return v1.ImageResolution_IMG_RESOLUTION_2K
	default:
		return v1.ImageResolution_IMG_RESOLUTION_1K
	}
}

// ImageRequest builds an image generation request.
type ImageRequest struct {
	prompt      string
	model       string
	n           *int32
	user        string
	format      *ImageFormat
	aspectRatio *ImageAspectRatio
	resolution  *ImageResolution
	inputImage  string
}

// NewImageRequest creates a new image generation request.
func NewImageRequest(prompt string) *ImageRequest {
	return &ImageRequest{prompt: prompt}
}

// WithModel sets the model to use.
func (r *ImageRequest) WithModel(model string) *ImageRequest {
	r.model = model
	return r
}

// WithCount sets the number of images to generate (1-10).
func (r *ImageRequest) WithCount(n int32) *ImageRequest {
	r.n = &n
	return r
}

// WithUser sets an opaque user identifier.
func (r *ImageRequest) WithUser(user string) *ImageRequest {
	r.user = user
	return r
}

// WithFormat sets the output image format.
func (r *ImageRequest) WithFormat(format ImageFormat) *ImageRequest {
	r.format = &format
	return r
}

// WithAspectRatio sets the image aspect ratio.
func (r *ImageRequest) WithAspectRatio(ar ImageAspectRatio) *ImageRequest {
	r.aspectRatio = &ar
	return r
}

// WithResolution sets the image resolution.
func (r *ImageRequest) WithResolution(res ImageResolution) *ImageRequest {
	r.resolution = &res
	return r
}

// WithInputImage sets an input image URL for image-to-image generation.
func (r *ImageRequest) WithInputImage(url string) *ImageRequest {
	r.inputImage = url
	return r
}

func (r *ImageRequest) toProto() *v1.GenerateImageRequest {
	req := &v1.GenerateImageRequest{
		Prompt: r.prompt,
		Model:  r.model,
		User:   r.user,
		Format: v1.ImageFormat_IMG_FORMAT_URL, // Default to URL format
	}
	if r.n != nil {
		req.N = r.n
	}
	if r.format != nil {
		req.Format = r.format.toProto()
	}
	if r.aspectRatio != nil {
		ar := r.aspectRatio.toProto()
		req.AspectRatio = &ar
	}
	if r.resolution != nil {
		res := r.resolution.toProto()
		req.Resolution = &res
	}
	if r.inputImage != "" {
		req.Image = &v1.ImageUrlContent{
			ImageUrl: r.inputImage,
		}
	}
	return req
}

// GeneratedImage represents a generated image.
type GeneratedImage struct {
	// URL is the URL where the image can be downloaded.
	URL string
	// Base64 is the base64-encoded image data (if requested).
	Base64 string
	// RespectModeration indicates if the image respects moderation rules.
	RespectModeration bool
}

// ImageResponse contains the image generation results.
type ImageResponse struct {
	// Images are the generated images.
	Images []GeneratedImage
	// Model is the model that was used.
	Model string
}

// GenerateImage generates images from a text prompt.
func (c *Client) GenerateImage(ctx context.Context, req *ImageRequest) (*ImageResponse, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	resp, err := c.image.GenerateImage(ctx, req.toProto())
	if err != nil {
		return nil, FromGRPCError(err)
	}

	result := &ImageResponse{
		Model: resp.GetModel(),
	}

	for _, img := range resp.GetImages() {
		gi := GeneratedImage{
			RespectModeration: img.GetRespectModeration(),
		}
		if img.GetUrl() != "" {
			gi.URL = img.GetUrl()
		}
		if img.GetBase64() != "" {
			gi.Base64 = img.GetBase64()
		}
		result.Images = append(result.Images, gi)
	}

	return result, nil
}
