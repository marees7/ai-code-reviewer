package cost

const tokensPer1K = 1000.0

type ModelPrice struct {
	InputPer1KUSD  float64
	OutputPer1KUSD float64
}

var prices = map[string]ModelPrice{
	// Update these constants as provider pricing changes.
	"gpt-3.5-turbo": {InputPer1KUSD: 0.0005, OutputPer1KUSD: 0.0015},
	"gpt-4o-mini":   {InputPer1KUSD: 0.00015, OutputPer1KUSD: 0.0006},
	"gpt-4o":        {InputPer1KUSD: 0.005, OutputPer1KUSD: 0.015},
}

func EstimateUSD(model string, promptTokens, completionTokens int) float64 {
	price, ok := prices[model]
	if !ok {
		return 0
	}

	inputCost := (float64(promptTokens) / tokensPer1K) * price.InputPer1KUSD
	outputCost := (float64(completionTokens) / tokensPer1K) * price.OutputPer1KUSD
	return inputCost + outputCost
}
