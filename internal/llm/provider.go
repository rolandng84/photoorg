package llm

import "context"

// VisionProvider is the interface for LLM providers that can categorize images
type VisionProvider interface {
	CategorizeImage(ctx context.Context, imageBytes []byte, categories []string, model string, customPrompt string) (string, error)
	ListModels(ctx context.Context) ([]string, error)
}

// DefaultPrompt generates the categorization prompt
func DefaultPrompt(categories []string, customPrompt string) string {
	catStr := ""
	for i, c := range categories {
		if i > 0 {
			catStr += ", "
		}
		catStr += c
	}

	if customPrompt != "" {
		// Replace [categories] placeholder
		result := customPrompt
		for {
			idx := indexOf(result, "[categories]")
			if idx == -1 {
				break
			}
			result = result[:idx] + catStr + result[idx+len("[categories]"):]
		}
		return result
	}

	return "IMPORTANT RULE: If ANY person or human is visible in this image, you MUST respond with \"people\". This rule overrides everything else.\n\nNow categorize this image into exactly one category: " + catStr + ".\n\nRemember: humans visible = \"people\", no exceptions. Respond with only the category name, nothing else."
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
