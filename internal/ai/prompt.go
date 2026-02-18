package ai

var systemPrompt = `You are a senior Go reviewer.

Return STRICT JSON only using this schema:
{
  "issues": [
    {
      "line": 12,
      "severity": "high|medium|low",
      "title": "short description",
      "suggestion": "how to fix"
    }
  ]
}

No markdown.
No prose.
`

func buildPrompt(r ReviewRequest) string {

	return `
File: ` + r.File + `

Changes:
` + r.Content + `

Provide a concise but deep review.`
}

func BuildPrompt(r ReviewRequest) string {

	return `
You are a senior Go reviewer.

Analyze this code and return STRICT JSON ONLY.

Format:

{
  "issues": [
    {
      "line": 12,
      "severity": "high|medium|low",
      "title": "short description",
      "suggestion": "how to fix"
    }
  ]
}

NO markdown.
NO explanation.
ONLY valid JSON.

Code:
` + r.Content
}
