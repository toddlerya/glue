package system

import (
	"os"
	"runtime"
)

// IsWindows check if current os is windows.
// Play: https://go.dev/play/p/XzJULbzmf9m
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsLinux check if current os is linux.
// Play: https://go.dev/play/p/zIflQgZNuxD
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// IsMac check if current os is macos.
// Play: https://go.dev/play/p/Mg4Hjtyq7Zc
func IsMac() bool {
	return runtime.GOOS == "darwin"
}

// GetOsEnv gets the value of the environment variable named by the key.
// Play: https://go.dev/play/p/D88OYVCyjO-
func GetOsEnv(key string) string {
	return os.Getenv(key)
}

// SetOsEnv sets the value of the environment variable named by the key.
// Play: https://go.dev/play/p/D88OYVCyjO-
func SetOsEnv(key, value string) error {
	return os.Setenv(key, value)
}

// RemoveOsEnv remove a single environment variable.
// Play: https://go.dev/play/p/fqyq4b3xUFQ
func RemoveOsEnv(key string) error {
	return os.Unsetenv(key)
}

// CompareOsEnv gets env named by the key and compare it with comparedEnv.
// Play: https://go.dev/play/p/BciHrKYOHbp
func CompareOsEnv(key, comparedEnv string) bool {
	env := GetOsEnv(key)
	if env == "" {
		return false
	}
	return env == comparedEnv
}

// GetOsBits return current os bits (32 or 64).
// Play: https://go.dev/play/p/ml-_XH3gJbW
func GetOsBits() int {
	return 32 << (^uint(0) >> 63)
}
