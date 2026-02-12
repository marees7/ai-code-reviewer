package ai

import "ai-code-reviewer/internal/config"

func NewProvider(cfg *config.Config) Provider {

	switch cfg.AIProvider {

	case "ollama":
		return NewOllama(
			cfg.OllamaURL,
			cfg.OllamaModel,
		)

	default:
		return NewOpenAI(
			cfg.OpenAIKey,
			cfg.OpenAIModel,
		)
	}
}
