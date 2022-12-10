package generators

import (
	"create-cli/internal/log"
	"crypto/rand"
	"encoding/hex"
	"os"

	"github.com/sethvargo/go-password/password"
)

var logger = log.StderrLogger{Stderr: os.Stderr, Tool: "Secret generator"}

func GenerateSecureSecret(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func GenerateSecret(length int, numDigits int, numSymbls int, noUpper bool) string {
	// Generate a password that is 64 characters long with 10 digits, 10 symbols,
	// allowing upper and lower case letters, disallowing repeat characters.
	logger.Waitingf("Generating secret...")
	res, err := password.Generate(length, numDigits, numSymbls, noUpper, false)
	if err != nil {
		logger.Failuref("Error generating secret", err)
	}

	logger.Successf("Secret generated")
	return res
}
