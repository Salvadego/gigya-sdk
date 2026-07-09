package gigya

import "encoding/json"

func jsonUnmarshalToken(body []byte, out *TokenResponse) error {
	return json.Unmarshal(body, out)
}
