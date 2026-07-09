package gigya

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

func signParams(endpointURL string, params url.Values, userKey, secret string) error {
	if userKey == "" || secret == "" {
		return fmt.Errorf("classic auth requires UserKey and Secret")
	}

	params.Set("userKey", userKey)
	params.Set("timestamp", strconv.FormatInt(time.Now().Unix(), 10))
	params.Set("nonce", nonce())

	baseString := "POST&" + rfc3986Escape(endpointURL) + "&" + rfc3986Escape(sortedEncode(params))

	key, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return fmt.Errorf("secret is not valid base64: %w", err)
	}

	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(baseString))
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	params.Set("sig", sig)
	return nil
}

func sortedEncode(params url.Values) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte('&')
		}
		b.WriteString(rfc3986Escape(k))
		b.WriteByte('=')
		b.WriteString(rfc3986Escape(params.Get(k)))
	}
	return b.String()
}

func rfc3986Escape(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if isUnreserved(c) {
			b.WriteByte(c)
		} else {
			fmt.Fprintf(&b, "%%%02X", c)
		}
	}
	return b.String()
}

func isUnreserved(c byte) bool {
	switch {
	case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c >= '0' && c <= '9':
		return true
	case c == '-' || c == '_' || c == '.' || c == '~':
		return true
	default:
		return false
	}
}

func nonce() string {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 16)
	for i := range b {
		b[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return fmt.Sprintf("%d_%s", time.Now().UnixNano(), string(b))
}
