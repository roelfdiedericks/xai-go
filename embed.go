package xai

import (
	"context"

	v1 "github.com/roelfdiedericks/xai-go/proto/xai/api/v1"
)

// EmbedInput represents an input to embed.
type EmbedInput interface {
	isEmbedInput()
	toProto() *v1.EmbedInput
}

// TextEmbedInput is a text input for embedding.
type TextEmbedInput struct {
	Text string
}

func (TextEmbedInput) isEmbedInput() {}

func (t TextEmbedInput) toProto() *v1.EmbedInput {
	return &v1.EmbedInput{
		Input: &v1.EmbedInput_String_{String_: t.Text},
	}
}

// ImageEmbedInput is an image URL input for embedding.
type ImageEmbedInput struct {
	URL string
}

func (ImageEmbedInput) isEmbedInput() {}

func (i ImageEmbedInput) toProto() *v1.EmbedInput {
	return &v1.EmbedInput{
		Input: &v1.EmbedInput_ImageUrl{
			ImageUrl: &v1.ImageUrlContent{ImageUrl: i.URL},
		},
	}
}

// EmbedRequest builds an embedding request.
type EmbedRequest struct {
	model  string
	inputs []EmbedInput
	user   string
}

// NewEmbedRequest creates a new embedding request for the specified model.
func NewEmbedRequest(model string) *EmbedRequest {
	return &EmbedRequest{model: model}
}

// AddText adds a text input to embed.
func (r *EmbedRequest) AddText(text string) *EmbedRequest {
	r.inputs = append(r.inputs, TextEmbedInput{Text: text})
	return r
}

// AddTexts adds multiple text inputs to embed.
func (r *EmbedRequest) AddTexts(texts ...string) *EmbedRequest {
	for _, t := range texts {
		r.inputs = append(r.inputs, TextEmbedInput{Text: t})
	}
	return r
}

// AddImage adds an image URL input to embed.
func (r *EmbedRequest) AddImage(url string) *EmbedRequest {
	r.inputs = append(r.inputs, ImageEmbedInput{URL: url})
	return r
}

// WithUser sets an opaque user identifier for logging.
func (r *EmbedRequest) WithUser(user string) *EmbedRequest {
	r.user = user
	return r
}

func (r *EmbedRequest) toProto() *v1.EmbedRequest {
	req := &v1.EmbedRequest{
		Model: r.model,
		User:  r.user,
	}
	for _, input := range r.inputs {
		req.Input = append(req.Input, input.toProto())
	}
	return req
}

// Embedding represents a single embedding result.
type Embedding struct {
	// Index is the position of this embedding in the request.
	Index int32
	// Vectors are the embedding vectors (may have multiple for images).
	Vectors [][]float32
}

// EmbedResponse contains the embedding results.
type EmbedResponse struct {
	// Embeddings are the generated embeddings.
	Embeddings []Embedding
	// Model is the model that was used.
	Model string
	// NumTextEmbeddings is the number of text embeddings generated.
	NumTextEmbeddings int32
	// NumImageEmbeddings is the number of image embeddings generated.
	NumImageEmbeddings int32
}

// Embed generates embeddings for the given inputs.
func (c *Client) Embed(ctx context.Context, req *EmbedRequest) (*EmbedResponse, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	resp, err := c.embedder.Embed(ctx, req.toProto())
	if err != nil {
		return nil, FromGRPCError(err)
	}

	result := &EmbedResponse{
		Model: resp.GetModel(),
	}

	if usage := resp.GetUsage(); usage != nil {
		result.NumTextEmbeddings = usage.GetNumTextEmbeddings()
		result.NumImageEmbeddings = usage.GetNumImageEmbeddings()
	}

	for _, emb := range resp.GetEmbeddings() {
		embedding := Embedding{
			Index: emb.GetIndex(),
		}
		for _, fv := range emb.GetEmbeddings() {
			embedding.Vectors = append(embedding.Vectors, fv.GetFloatArray())
		}
		result.Embeddings = append(result.Embeddings, embedding)
	}

	return result, nil
}
