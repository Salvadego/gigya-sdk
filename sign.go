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

	baseString := "POST&" + url.QueryEscape(endpointURL) + "&" + url.QueryEscape(sortedEncode(params))

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
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(url.QueryEscape(params.Get(k)))
	}
	return b.String()
}

func nonce() string {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 16)
	for i := range b {
		b[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return fmt.Sprintf("%d_%s", time.Now().UnixNano(), string(b))
}
