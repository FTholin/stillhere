package api

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/FTholin/stillhere/internal/switches"
)

// newToken returns a 256 bit URL safe random token
func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

// newID returns a 128-bit identifier. It is random, not sequential:
// knowing a switch's ID must not let you guess the next one.
func newID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// identify assigns the server-owned identifiers. It is the only place
// where a Switch stops being anonymous
func identify(sw *switches.Switch) error {
	var err error

	if sw.ID, err = newID(); err != nil {
		return fmt.Errorf("new id: %w", err)
	}

	if sw.CheckInToken, err = newToken(); err != nil {
		return fmt.Errorf("new check-in token: %w", err)
	}

	if sw.RevealToken, err = newToken(); err != nil {
		return fmt.Errorf("new reveal token: %w", err)
	}

	return nil
}
