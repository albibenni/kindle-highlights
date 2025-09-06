package types

import (
	"os"
	"runtime"
)

type EnvVar string

const (
	NotePath     EnvVar = "NOTE_PATH"
	ClippingPath EnvVar = "CLIPPING_PATH"
)

type EnvFile string

const (
	Mac   EnvFile = "mac.env"
	Linux EnvFile = "linux.env"
)

func (e EnvVar) Value() string {
	return os.Getenv(string(e))
}

func (e EnvVar) String() string {
	return string(e)
}

func (e EnvFile) Value() string {
	return os.Getenv(string(e))
}

func (e EnvFile) String() string {
	return string(e)
}

func GetEnvFile() string {
	switch runtime.GOOS {
	case "darwin":
		return string(Mac)
	case "linux":
		return string(Linux)
	case "windows":
		return "wrong pc"
	default:
		return string(Mac) // default fallback
	}
}

