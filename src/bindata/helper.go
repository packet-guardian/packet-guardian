package bindata

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

var customDir = "custom"

// SetCustomDir sets the directory that contains files to override
// builtin templates or other static assets.
func SetCustomDir(dir string) error {
	stat, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		return errors.New("Custom directory must be a directory")
	}

	customDir = dir
	return nil
}

// GetAsset wraps go-bindata's Asset function and checks if a file
// with the same path/name exists in the custom directory it will
// be loaded and used instead of the builtin asset. Otherwise, the
// builtin asset will be returned.
func GetAsset(name string) ([]byte, error) {
	full, err := ioutil.ReadFile(filepath.Join(customDir, name))
	if err == nil { // File exists in custom directory, use it
		return full, nil
	}

	return Asset(name)
}

// GetAssetInfo wraps go-bindata's AssetInfo function and checks
// if a file with the same path/name exists in the custom directory.
// If a custom file exists, it will be stat-ed instead. If it doesn't
// exist, the default AssetInfo function is called and returned.
func GetAssetInfo(name string) (os.FileInfo, error) {
	stat, err := os.Stat(filepath.Join(customDir, name))
	if err == nil {
		return stat, nil
	}

	return AssetInfo(name)
}

// GetAssetDir wraps go-bindata's AssetDir function. It merges the
// bindata with anything in the custom folder that can be accessed
// using GetAsset or Asset.
func GetAssetDir(name string) ([]string, error) {
	return AssetDir(name)
}
