package config

import (
	"crypto/md5" // #nosec G501 -- used only for stable legacy database IDs, not cryptography
	"encoding/hex"
)

func GetMD5Hash(text string) string {
	hasher := md5.New() // #nosec G401 -- changing this legacy identifier would duplicate existing records
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
