package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

var (
	adjectives = []string{
		"swift", "bright", "clever", "happy", "quick", "bold", "calm", "deep", "fast", "great",
		"kind", "light", "new", "old", "proud", "quiet", "rich", "safe", "tall", "warm",
		"wise", "young", "big", "small", "strong", "weak", "hot", "cold", "dry", "wet",
		"hard", "soft", "loud", "silent", "fresh", "stale", "clean", "dirty", "smooth", "rough",
		"sharp", "dull", "thick", "thin", "wide", "narrow", "long", "short", "high", "low",
		"independent", "reliable", "efficient", "modern", "classic", "stable", "dynamic", "secure",
	}

	nouns = []string{
		"server", "engine", "cloud", "node", "host", "box", "core", "hub", "lab", "desk",
		"tower", "bridge", "gate", "port", "net", "web", "link", "zone", "base", "unit",
		"system", "device", "machine", "platform", "service", "cluster", "instance", "worker",
		"runner", "builder", "manager", "handler", "monitor", "tracker", "scanner", "parser",
		"router", "proxy", "cache", "store", "vault", "shield", "guard", "watch", "beacon", "signal",
	}
)

// GeneratePassword generates a random alphanumeric password of the specified length
func GeneratePassword(length int) (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)

	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}

	return string(result), nil
}

// GenerateRandomVMName generates a random VM name in the format: adjective-noun-number
func GenerateRandomVMName() (string, error) {
	adjIdx, err := rand.Int(rand.Reader, big.NewInt(int64(len(adjectives))))
	if err != nil {
		return "", err
	}

	nounIdx, err := rand.Int(rand.Reader, big.NewInt(int64(len(nouns))))
	if err != nil {
		return "", err
	}

	number, err := rand.Int(rand.Reader, big.NewInt(900))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%s-%d",
		adjectives[adjIdx.Int64()],
		nouns[nounIdx.Int64()],
		number.Int64()+100,
	), nil
}
