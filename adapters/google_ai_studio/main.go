package google_ai_studio

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

type GoogleStudioAdapter struct {
	client *genai.Client
	model  string
	config *genai.GenerateContentConfig
}

func NewGoogleStudioAdapter(apiKey string, model string) (*GoogleStudioAdapter, error) {
	ctx := context.Background()
	clientConfig := &genai.ClientConfig{
		APIKey: apiKey,
	}

	client, err := genai.NewClient(ctx, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Google AI Studio client: %w", err)
	}

	if model == "" {
		model = "gemini-2.5-flash"
	}

	thinking_budget := int32(0) // Disables thinking
	// TODO: add values for temperature and other configs. Check the docs.
	config := &genai.GenerateContentConfig{
		ThinkingConfig: &genai.ThinkingConfig{
			ThinkingBudget: &thinking_budget,
		},
	}
	return &GoogleStudioAdapter{
		client: client,
		model:  model,
		config: config,
	}, nil
}

func (g *GoogleStudioAdapter) GenerateText(ctx context.Context, input string) (string, error) {
	parts := []*genai.Part{
		genai.NewPartFromText(input),
	}

	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}

	result, err := g.client.Models.GenerateContent(
		ctx,
		g.model,
		contents,
		g.config,
	)

	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	return result.Text(), nil
}

func (g *GoogleStudioAdapter) GenerateTextWithImage(ctx context.Context, input string, imageData []byte, mimeType string) (string, error) {
	parts := []*genai.Part{
		genai.NewPartFromBytes(imageData, mimeType),
		genai.NewPartFromText(input),
	}

	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}

	result, err := g.client.Models.GenerateContent(
		ctx,
		g.model,
		contents,
		g.config,
	)

	if err != nil {
		return "", fmt.Errorf("failed to generate content with image: %w", err)
	}

	return result.Text(), nil
}
