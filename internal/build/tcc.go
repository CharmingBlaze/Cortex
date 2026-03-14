// Package tcc provides TCC (Tiny C Compiler) integration for Cortex
// TCC is a small, fast C compiler that can be bundled with Cortex
package build

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const (
	// TCC download URLs for different platforms
	tccWindowsURL = "https://github.com/TinyCC/tinycc/releases/download/tcc-0.9.27/tcc-0.9.27-win64-bin.zip"
	tccLinuxURL   = "https://github.com/TinyCC/tinycc/releases/download/tcc-0.9.27/tcc-0.9.27-x86_64.tar.gz"
	tccMacURL     = "https://github.com/TinyCC/tinycc/releases/download/tcc-0.9.27/tcc-0.9.27-x86_64-apple-darwin11.tar.gz"
)

// BundledTCC manages a bundled TCC compiler
type BundledTCC struct {
	InstallPath string
}

// FindOrInstall locates TCC or installs it if not found
func FindOrInstall() (*BundledTCC, error) {
	// First, check if TCC is already in PATH
	if path, err := exec.LookPath("tcc"); err == nil {
		return &BundledTCC{InstallPath: filepath.Dir(path)}, nil
	}

	// Check for bundled TCC in cortex directory
	cortexDir := getTccCortexDir()
	tccDir := filepath.Join(cortexDir, "tcc")

	if tccExe := FindTCCExecutable(tccDir); tccExe != "" {
		return &BundledTCC{InstallPath: tccDir}, nil
	}

	// Download and install TCC
	fmt.Println("TCC not found. Downloading and installing...")
	if err := DownloadAndInstall(tccDir); err != nil {
		return nil, fmt.Errorf("failed to install TCC: %w", err)
	}

	return &BundledTCC{InstallPath: tccDir}, nil
}

// getCortexDir returns the directory where cortex is installed
func getTccCortexDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

// findTCCExecutable locates the TCC executable in the given directory
func FindTCCExecutable(tccDir string) string {
	exeName := "tcc"
	if runtime.GOOS == "windows" {
		exeName = "tcc.exe"
	}

	// Check common subdirectories
	paths := []string{
		filepath.Join(tccDir, exeName),
		filepath.Join(tccDir, "bin", exeName),
		filepath.Join(tccDir, "x86_64-win64-tcc", exeName),
		filepath.Join(tccDir, "x86_64-linux-tcc", exeName),
		filepath.Join(tccDir, "x86_64-osx-tcc", exeName),
	}

	for _, path := range paths {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}

	return ""
}

// downloadAndInstall downloads and installs TCC
func DownloadAndInstall(tccDir string) error {
	if err := os.MkdirAll(tccDir, 0755); err != nil {
		return err
	}

	url := GetTCCDownloadURL()

	fmt.Printf("Downloading TCC from %s...\n", url)

	// Download archive
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	// Save to temporary file
	tmpFile := filepath.Join(os.TempDir(), "tcc-download")
	if runtime.GOOS == "windows" {
		tmpFile += ".zip"
	} else {
		tmpFile += ".tar.gz"
	}

	out, err := os.Create(tmpFile)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		return err
	}
	out.Close()

	// Extract archive
	fmt.Println("Extracting TCC...")
	if runtime.GOOS == "windows" {
		err = ExtractZip(tmpFile, tccDir)
	} else {
		err = ExtractTarGz(tmpFile, tccDir)
	}

	os.Remove(tmpFile)

	if err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	fmt.Println("TCC installed successfully!")
	return nil
}

// getTCCDownloadURL returns the appropriate download URL for the current platform
func GetTCCDownloadURL() string {
	switch runtime.GOOS {
	case "windows":
		return tccWindowsURL
	case "darwin":
		return tccMacURL
	default:
		return tccLinuxURL
	}
}

// extractZip extracts a ZIP archive
func ExtractZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// extractTarGz extracts a tar.gz archive
func ExtractTarGz(src, dest string) error {
	// This is a simplified implementation
	// In production, use proper tar extraction
	cmd := exec.Command("tar", "-xzf", src, "-C", dest)
	return cmd.Run()
}

// GetCompilerPath returns the path to the TCC compiler
func (t *BundledTCC) GetCompilerPath() string {
	return FindTCCExecutable(t.InstallPath)
}

// Compile compiles a C source file using TCC
func (t *BundledTCC) Compile(source, output string, includes []string) error {
	tccPath := t.GetCompilerPath()
	if tccPath == "" {
		return fmt.Errorf("TCC executable not found")
	}

	args := []string{"-c", source, "-o", output}
	for _, inc := range includes {
		args = append(args, "-I", inc)
	}

	cmd := exec.Command(tccPath, args...)
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("compilation failed: %s\n%s", err, string(outputBytes))
	}

	return nil
}

// Link links object files into an executable using TCC
func (t *BundledTCC) Link(objects []string, output string, libraries []string) error {
	tccPath := t.GetCompilerPath()
	if tccPath == "" {
		return fmt.Errorf("TCC executable not found")
	}

	args := append([]string{"-o", output}, objects...)
	for _, lib := range libraries {
		args = append(args, "-l", lib)
	}

	cmd := exec.Command(tccPath, args...)
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("linking failed: %s\n%s", err, string(outputBytes))
	}

	return nil
}
