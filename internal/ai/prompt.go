package ai

var systemPrompt = `You are a senior Go reviewer.

Respond ONLY in this format:

LINE: <number>: <issue explanation>

If fix obvious include:

SUGGESTION:
` + "```go\n<code>\n```" + `
`

func buildPrompt(r ReviewRequest) string {

	return `
File: ` + r.File + `

Changes:
` + r.Content + `

Provide a concise but deep review.`
}
