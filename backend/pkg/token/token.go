package token

import "github.com/pkoukk/tiktoken-go"

func CountTokens(text string) int {
	enc, err := tiktoken.EncodingForModel("gpt-3.5-turbo")
	if err != nil {
		enc, _ = tiktoken.GetEncoding("cl100k_base")
	}
	return len(enc.Encode(text, nil, nil))
}
