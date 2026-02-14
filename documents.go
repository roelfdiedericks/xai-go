package xai

import (
	"context"

	v1 "github.com/roelfdiedericks/xai-go/proto/xai/api/v1"
)

// RetrievalMode specifies how to perform document search.
type RetrievalMode int

const (
	// RetrievalModeHybrid uses both semantic and keyword search.
	RetrievalModeHybrid RetrievalMode = iota
	// RetrievalModeSemantic uses only semantic (embedding) search.
	RetrievalModeSemantic
	// RetrievalModeKeyword uses only keyword search.
	RetrievalModeKeyword
)

// SearchRequest builds a document search request.
type SearchRequest struct {
	query         string
	collectionIDs []string
	limit         *int32
	instructions  *string
	mode          *RetrievalMode
}

// NewSearchRequest creates a new document search request.
func NewSearchRequest(query string) *SearchRequest {
	return &SearchRequest{query: query}
}

// WithCollections specifies which collections to search.
func (r *SearchRequest) WithCollections(ids ...string) *SearchRequest {
	r.collectionIDs = ids
	return r
}

// WithLimit sets the maximum number of chunks to return.
func (r *SearchRequest) WithLimit(n int32) *SearchRequest {
	r.limit = &n
	return r
}

// WithInstructions sets custom search instructions.
func (r *SearchRequest) WithInstructions(instructions string) *SearchRequest {
	r.instructions = &instructions
	return r
}

// WithRetrievalMode sets the retrieval mode.
func (r *SearchRequest) WithRetrievalMode(mode RetrievalMode) *SearchRequest {
	r.mode = &mode
	return r
}

func (r *SearchRequest) toProto() *v1.SearchRequest {
	req := &v1.SearchRequest{
		Query: r.query,
		Source: &v1.DocumentsSource{
			CollectionIds: r.collectionIDs,
		},
	}
	if r.limit != nil {
		req.Limit = r.limit
	}
	if r.instructions != nil {
		req.Instructions = r.instructions
	}
	if r.mode != nil {
		switch *r.mode {
		case RetrievalModeSemantic:
			req.RetrievalMode = &v1.SearchRequest_SemanticRetrieval{
				SemanticRetrieval: &v1.SemanticRetrieval{},
			}
		case RetrievalModeKeyword:
			req.RetrievalMode = &v1.SearchRequest_KeywordRetrieval{
				KeywordRetrieval: &v1.KeywordRetrieval{},
			}
		default:
			req.RetrievalMode = &v1.SearchRequest_HybridRetrieval{
				HybridRetrieval: &v1.HybridRetrieval{},
			}
		}
	}
	return req
}

// SearchMatch represents a matching document chunk.
type SearchMatch struct {
	// Content is the chunk content.
	Content string
	// Score is the relevance score.
	Score float32
	// FileID is the source file/document ID.
	FileID string
	// ChunkID is the chunk identifier.
	ChunkID string
	// CollectionIDs are the collections this document belongs to.
	CollectionIDs []string
}

// SearchResponse contains the document search results.
type SearchResponse struct {
	// Matches are the matching document chunks.
	Matches []SearchMatch
}

// SearchDocuments searches document collections.
func (c *Client) SearchDocuments(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	resp, err := c.documents.Search(ctx, req.toProto())
	if err != nil {
		return nil, FromGRPCError(err)
	}

	result := &SearchResponse{}
	for _, match := range resp.GetMatches() {
		result.Matches = append(result.Matches, SearchMatch{
			Content:       match.GetChunkContent(),
			Score:         match.GetScore(),
			FileID:        match.GetFileId(),
			ChunkID:       match.GetChunkId(),
			CollectionIDs: match.GetCollectionIds(),
		})
	}

	return result, nil
}
