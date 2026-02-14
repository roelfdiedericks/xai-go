package xai

import (
	"context"
	"time"

	v1 "github.com/roelfdiedericks/xai-go/proto/xai/api/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Modality represents input/output modality supported by a model.
type Modality string

const (
	// ModalityText represents text input/output.
	ModalityText Modality = "text"
	// ModalityImage represents image input/output.
	ModalityImage Modality = "image"
	// ModalityEmbedding represents embedding output.
	ModalityEmbedding Modality = "embedding"
)

func modalityFromProto(m v1.Modality) Modality {
	switch m {
	case v1.Modality_TEXT:
		return ModalityText
	case v1.Modality_IMAGE:
		return ModalityImage
	case v1.Modality_EMBEDDING:
		return ModalityEmbedding
	default:
		return ""
	}
}

// Pricing represents token pricing in USD.
type Pricing struct {
	// PerMillionTokens is the price per million tokens in USD.
	PerMillionTokens float64
}

// LanguageModel represents a language model available on the platform.
type LanguageModel struct {
	// Name is the model name used in API requests.
	Name string
	// Aliases are alternative names for the model.
	Aliases []string
	// Version is the model version number.
	Version string
	// InputModalities are the supported input types.
	InputModalities []Modality
	// OutputModalities are the supported output types.
	OutputModalities []Modality
	// MaxPromptLength is the maximum context length (tokens).
	MaxPromptLength int32
	// Created is when the model was created.
	Created time.Time
	// SystemFingerprint identifies the model configuration.
	SystemFingerprint string
	// PromptTextPricing is the price for prompt text tokens.
	PromptTextPricing Pricing
	// PromptImagePricing is the price for prompt image tokens.
	PromptImagePricing Pricing
	// CachedPromptPricing is the price for cached prompt tokens.
	CachedPromptPricing Pricing
	// CompletionPricing is the price for completion tokens.
	CompletionPricing Pricing
	// SearchPricing is the price per search.
	SearchPricing Pricing
}

// SupportsImages returns true if the model can process image inputs.
func (m *LanguageModel) SupportsImages() bool {
	for _, mod := range m.InputModalities {
		if mod == ModalityImage {
			return true
		}
	}
	return false
}

// CalculateCost estimates the cost in USD for a request.
func (m *LanguageModel) CalculateCost(inputTokens, outputTokens, cacheReadTokens int) float64 {
	inputCost := float64(inputTokens) * m.PromptTextPricing.PerMillionTokens / 1_000_000
	outputCost := float64(outputTokens) * m.CompletionPricing.PerMillionTokens / 1_000_000
	cacheCost := float64(cacheReadTokens) * m.CachedPromptPricing.PerMillionTokens / 1_000_000
	return inputCost + outputCost + cacheCost
}

func languageModelFromProto(m *v1.LanguageModel) *LanguageModel {
	model := &LanguageModel{
		Name:              m.GetName(),
		Aliases:           m.GetAliases(),
		Version:           m.GetVersion(),
		MaxPromptLength:   m.GetMaxPromptLength(),
		SystemFingerprint: m.GetSystemFingerprint(),
	}

	if m.GetCreated() != nil {
		model.Created = m.GetCreated().AsTime()
	}

	for _, mod := range m.GetInputModalities() {
		if mm := modalityFromProto(mod); mm != "" {
			model.InputModalities = append(model.InputModalities, mm)
		}
	}

	for _, mod := range m.GetOutputModalities() {
		if mm := modalityFromProto(mod); mm != "" {
			model.OutputModalities = append(model.OutputModalities, mm)
		}
	}

	// Pricing: convert from 1/100 USD cents per million to USD per million
	// Price in proto is in 1/100 cents, so divide by 10000 to get dollars
	model.PromptTextPricing = Pricing{PerMillionTokens: float64(m.GetPromptTextTokenPrice()) / 10000}
	model.PromptImagePricing = Pricing{PerMillionTokens: float64(m.GetPromptImageTokenPrice()) / 10000}
	model.CachedPromptPricing = Pricing{PerMillionTokens: float64(m.GetCachedPromptTokenPrice()) / 10000}
	model.CompletionPricing = Pricing{PerMillionTokens: float64(m.GetCompletionTextTokenPrice()) / 10000}
	model.SearchPricing = Pricing{PerMillionTokens: float64(m.GetSearchPrice()) / 10000}

	return model
}

// ListModels returns all available language models.
func (c *Client) ListModels(ctx context.Context) ([]*LanguageModel, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	resp, err := c.models.ListLanguageModels(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, FromGRPCError(err)
	}

	models := make([]*LanguageModel, 0, len(resp.GetModels()))
	for _, m := range resp.GetModels() {
		models = append(models, languageModelFromProto(m))
	}

	return models, nil
}

// GetModel retrieves details about a specific language model.
func (c *Client) GetModel(ctx context.Context, name string) (*LanguageModel, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	resp, err := c.models.GetLanguageModel(ctx, &v1.GetModelRequest{Name: name})
	if err != nil {
		return nil, FromGRPCError(err)
	}

	return languageModelFromProto(resp), nil
}

// EmbeddingModel represents an embedding model.
type EmbeddingModel struct {
	// Name is the model name.
	Name string
	// Aliases are alternative names.
	Aliases []string
	// Version is the model version.
	Version string
	// InputModalities are supported input types.
	InputModalities []Modality
	// OutputModalities are supported output types.
	OutputModalities []Modality
	// Created is when the model was created.
	Created time.Time
	// SystemFingerprint identifies the model configuration.
	SystemFingerprint string
	// PromptTextPricing is the price for text prompt tokens.
	PromptTextPricing Pricing
	// PromptImagePricing is the price for image prompt tokens.
	PromptImagePricing Pricing
}

func embeddingModelFromProto(m *v1.EmbeddingModel) *EmbeddingModel {
	model := &EmbeddingModel{
		Name:               m.GetName(),
		Aliases:            m.GetAliases(),
		Version:            m.GetVersion(),
		SystemFingerprint:  m.GetSystemFingerprint(),
		PromptTextPricing:  Pricing{PerMillionTokens: float64(m.GetPromptTextTokenPrice()) / 10000},
		PromptImagePricing: Pricing{PerMillionTokens: float64(m.GetPromptImageTokenPrice()) / 10000},
	}

	if m.GetCreated() != nil {
		model.Created = m.GetCreated().AsTime()
	}

	for _, mod := range m.GetInputModalities() {
		if mm := modalityFromProto(mod); mm != "" {
			model.InputModalities = append(model.InputModalities, mm)
		}
	}

	for _, mod := range m.GetOutputModalities() {
		if mm := modalityFromProto(mod); mm != "" {
			model.OutputModalities = append(model.OutputModalities, mm)
		}
	}

	return model
}

// ListEmbeddingModels returns all available embedding models.
func (c *Client) ListEmbeddingModels(ctx context.Context) ([]*EmbeddingModel, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	resp, err := c.models.ListEmbeddingModels(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, FromGRPCError(err)
	}

	models := make([]*EmbeddingModel, 0, len(resp.GetModels()))
	for _, m := range resp.GetModels() {
		models = append(models, embeddingModelFromProto(m))
	}

	return models, nil
}

// GetEmbeddingModel retrieves details about a specific embedding model.
func (c *Client) GetEmbeddingModel(ctx context.Context, name string) (*EmbeddingModel, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	resp, err := c.models.GetEmbeddingModel(ctx, &v1.GetModelRequest{Name: name})
	if err != nil {
		return nil, FromGRPCError(err)
	}

	return embeddingModelFromProto(resp), nil
}

// ImageModel represents an image generation model.
type ImageModel struct {
	// Name is the model name.
	Name string
	// Aliases are alternative names.
	Aliases []string
	// Version is the model version.
	Version string
	// InputModalities are supported input types.
	InputModalities []Modality
	// OutputModalities are supported output types.
	OutputModalities []Modality
	// MaxPromptLength is the maximum context length.
	MaxPromptLength int32
	// Created is when the model was created.
	Created time.Time
	// PricePerImage is the price per generated image in USD.
	PricePerImage float64
}

func imageModelFromProto(m *v1.ImageGenerationModel) *ImageModel {
	model := &ImageModel{
		Name:            m.GetName(),
		Aliases:         m.GetAliases(),
		Version:         m.GetVersion(),
		MaxPromptLength: m.GetMaxPromptLength(),
		PricePerImage:   float64(m.GetImagePrice()) / 10000,
	}

	if m.GetCreated() != nil {
		model.Created = m.GetCreated().AsTime()
	}

	for _, mod := range m.GetInputModalities() {
		if mm := modalityFromProto(mod); mm != "" {
			model.InputModalities = append(model.InputModalities, mm)
		}
	}

	for _, mod := range m.GetOutputModalities() {
		if mm := modalityFromProto(mod); mm != "" {
			model.OutputModalities = append(model.OutputModalities, mm)
		}
	}

	return model
}

// ListImageModels returns all available image generation models.
func (c *Client) ListImageModels(ctx context.Context) ([]*ImageModel, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	resp, err := c.models.ListImageGenerationModels(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, FromGRPCError(err)
	}

	models := make([]*ImageModel, 0, len(resp.GetModels()))
	for _, m := range resp.GetModels() {
		models = append(models, imageModelFromProto(m))
	}

	return models, nil
}

// GetImageModel retrieves details about a specific image generation model.
func (c *Client) GetImageModel(ctx context.Context, name string) (*ImageModel, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	resp, err := c.models.GetImageGenerationModel(ctx, &v1.GetModelRequest{Name: name})
	if err != nil {
		return nil, FromGRPCError(err)
	}

	return imageModelFromProto(resp), nil
}
