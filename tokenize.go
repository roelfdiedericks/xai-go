package xai

import (
	"context"

	v1 "github.com/roelfdiedericks/xai-go/proto/xai/api/v1"
)

// Token represents a single token from tokenization.
type Token struct {
	// TokenID is the numeric token identifier.
	TokenID uint32
	// StringToken is the string representation of the token.
	StringToken string
}

// TokenizeResponse contains the tokenization result.
type TokenizeResponse struct {
	// Tokens are the tokenized results.
	Tokens []Token
}

// TokenCount returns the total number of tokens.
func (r *TokenizeResponse) TokenCount() int {
	return len(r.Tokens)
}

// Tokenize tokenizes text using the specified model's tokenizer.
func (c *Client) Tokenize(ctx context.Context, model, text string) (*TokenizeResponse, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	resp, err := c.tokenizer.TokenizeText(ctx, &v1.TokenizeTextRequest{
		Model: model,
		Text:  text,
	})
	if err != nil {
		return nil, FromGRPCError(err)
	}

	result := &TokenizeResponse{}
	for _, t := range resp.GetTokens() {
		result.Tokens = append(result.Tokens, Token{
			TokenID:     t.GetTokenId(),
			StringToken: t.GetStringToken(),
		})
	}

	return result, nil
}

// TokenizeWithModel tokenizes text using the client's default model.
func (c *Client) TokenizeWithModel(ctx context.Context, text string) (*TokenizeResponse, error) {
	return c.Tokenize(ctx, c.config.DefaultModel, text)
}
