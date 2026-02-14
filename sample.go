package xai

import (
	"context"
	"io"

	v1 "github.com/roelfdiedericks/xai-go/proto/xai/api/v1"
)

// SampleRequest builds a text sampling request.
type SampleRequest struct {
	prompts     []string
	model       string
	maxTokens   *int32
	seed        *int32
	stop        []string
	temperature *float32
	topP        *float32
}

// NewSampleRequest creates a new sampling request.
func NewSampleRequest(model string) *SampleRequest {
	return &SampleRequest{model: model}
}

// AddPrompt adds a text prompt to sample from.
func (r *SampleRequest) AddPrompt(prompt string) *SampleRequest {
	r.prompts = append(r.prompts, prompt)
	return r
}

// AddPrompts adds multiple text prompts.
func (r *SampleRequest) AddPrompts(prompts ...string) *SampleRequest {
	r.prompts = append(r.prompts, prompts...)
	return r
}

// WithMaxTokens sets the maximum tokens to generate.
func (r *SampleRequest) WithMaxTokens(n int32) *SampleRequest {
	r.maxTokens = &n
	return r
}

// WithSeed sets a random seed for reproducibility.
func (r *SampleRequest) WithSeed(seed int32) *SampleRequest {
	r.seed = &seed
	return r
}

// WithStop sets stop sequences.
func (r *SampleRequest) WithStop(sequences ...string) *SampleRequest {
	r.stop = sequences
	return r
}

// WithTemperature sets the sampling temperature.
func (r *SampleRequest) WithTemperature(t float32) *SampleRequest {
	r.temperature = &t
	return r
}

// WithTopP sets the nucleus sampling parameter.
func (r *SampleRequest) WithTopP(p float32) *SampleRequest {
	r.topP = &p
	return r
}

func (r *SampleRequest) toProto() *v1.SampleTextRequest {
	req := &v1.SampleTextRequest{
		Prompt: r.prompts,
		Model:  r.model,
		Stop:   r.stop,
	}
	if r.maxTokens != nil {
		req.MaxTokens = r.maxTokens
	}
	if r.seed != nil {
		req.Seed = r.seed
	}
	if r.temperature != nil {
		req.Temperature = r.temperature
	}
	if r.topP != nil {
		req.TopP = r.topP
	}
	return req
}

// SampleOutput represents a single sample output.
type SampleOutput struct {
	// Text is the generated text.
	Text string
	// FinishReason indicates why generation stopped.
	FinishReason FinishReason
	// Index is the output index.
	Index int32
}

// SampleResponse contains the sampling results.
type SampleResponse struct {
	// Outputs are the generated samples.
	Outputs []SampleOutput
	// Model is the model that was used.
	Model string
	// Usage contains token usage information.
	Usage Usage
}

// SampleText performs a text sampling request.
func (c *Client) SampleText(ctx context.Context, req *SampleRequest) (*SampleResponse, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	resp, err := c.sampler.SampleText(ctx, req.toProto())
	if err != nil {
		return nil, FromGRPCError(err)
	}

	return sampleResponseFromProto(resp), nil
}

func sampleResponseFromProto(resp *v1.SampleTextResponse) *SampleResponse {
	result := &SampleResponse{
		Model: resp.GetModel(),
		Usage: usageFromProto(resp.GetUsage()),
	}

	for _, choice := range resp.GetChoices() {
		result.Outputs = append(result.Outputs, SampleOutput{
			Text:         choice.GetText(),
			FinishReason: finishReasonFromProto(choice.GetFinishReason()),
			Index:        choice.GetIndex(),
		})
	}

	return result
}

// SampleStream is an iterator for streaming sample responses.
type SampleStream struct {
	stream v1.Sample_SampleTextStreamingClient
	err    error
}

// Next returns the next sample chunk, or io.EOF when done.
func (s *SampleStream) Next() (*SampleResponse, error) {
	if s.err != nil {
		return nil, s.err
	}

	resp, err := s.stream.Recv()
	if err == io.EOF {
		return nil, io.EOF
	}
	if err != nil {
		s.err = FromGRPCError(err)
		return nil, s.err
	}

	return sampleResponseFromProto(resp), nil
}

// Close closes the stream.
func (s *SampleStream) Close() error {
	return nil
}

// Err returns any error that occurred during streaming.
func (s *SampleStream) Err() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
}

// SampleTextStream starts a streaming text sampling request.
func (c *Client) SampleTextStream(ctx context.Context, req *SampleRequest) (*SampleStream, error) {
	stream, err := c.sampler.SampleTextStreaming(ctx, req.toProto())
	if err != nil {
		return nil, FromGRPCError(err)
	}

	return &SampleStream{stream: stream}, nil
}
