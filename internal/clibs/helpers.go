package clibs

import (
	"os"
	"path/filepath"
)

// InjectHelperFiles copies helper C files for a library to the output directory if specified in the config.
func InjectHelperFiles(config *LibraryConfig, outputDir string) error {
	if config == nil || len(config.HelperFiles) == 0 {
		return nil
	}

	for _, helperPath := range config.HelperFiles {
		if _, err := os.Stat(helperPath); os.IsNotExist(err) {
			continue // Skip non-existent helper files
		}

		destPath := filepath.Join(outputDir, filepath.Base(helperPath))
		data, err := os.ReadFile(helperPath)
		if err != nil {
			return err
		}
		if err := os.WriteFile(destPath, data, 0644); err != nil {
			return err
		}
	}
	return nil
}
