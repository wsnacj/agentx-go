package telemetry

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	privateObservationDirectoryMode os.FileMode = 0o700
	privateObservationFileMode      os.FileMode = 0o600
)

func preparePrivateJSONLPath(rawPath string, kind string) (string, error) {
	trimmed := strings.TrimSpace(rawPath)
	if trimmed == "" {
		return "", fmt.Errorf("agentx/telemetry: %s path is required", strings.TrimSpace(kind))
	}
	absPath, err := filepath.Abs(trimmed)
	if err != nil {
		return "", err
	}
	directory := filepath.Dir(absPath)
	created := false
	info, err := os.Lstat(directory)
	if errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(directory, privateObservationDirectoryMode); err != nil {
			return "", err
		}
		created = true
		info, err = os.Lstat(directory)
	}
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return "", fmt.Errorf("agentx/telemetry: observation parent must be a real directory: %s", directory)
	}
	if created {
		if err := os.Chmod(directory, privateObservationDirectoryMode); err != nil {
			return "", err
		}
		info, err = os.Lstat(directory)
		if err != nil {
			return "", err
		}
		if info.Mode().Perm() != privateObservationDirectoryMode.Perm() {
			return "", fmt.Errorf("agentx/telemetry: observation directory mode verification failed: got %#o want %#o", info.Mode().Perm(), privateObservationDirectoryMode)
		}
	}
	if _, err := os.Lstat(absPath); err == nil {
		file, openErr := openPrivateJSONLAppend(absPath)
		if openErr != nil {
			return "", openErr
		}
		if closeErr := file.Close(); closeErr != nil {
			return "", closeErr
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	return absPath, nil
}

func openPrivateJSONLAppend(path string) (*os.File, error) {
	for attempt := 0; attempt < 2; attempt++ {
		info, err := os.Lstat(path)
		switch {
		case errors.Is(err, os.ErrNotExist):
			file, createErr := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_APPEND|os.O_WRONLY, privateObservationFileMode)
			if errors.Is(createErr, os.ErrExist) {
				continue
			}
			if createErr != nil {
				return nil, createErr
			}
			if err := secureOpenedPrivateJSONLFile(path, file); err != nil {
				_ = file.Close()
				return nil, err
			}
			return file, nil
		case err != nil:
			return nil, err
		case info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular():
			return nil, fmt.Errorf("agentx/telemetry: observation path must be a regular file: %s", path)
		default:
			file, openErr := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, privateObservationFileMode)
			if errors.Is(openErr, os.ErrNotExist) {
				continue
			}
			if openErr != nil {
				return nil, openErr
			}
			if err := secureOpenedPrivateJSONLFile(path, file); err != nil {
				_ = file.Close()
				return nil, err
			}
			return file, nil
		}
	}
	return nil, fmt.Errorf("agentx/telemetry: observation path changed while opening: %s", path)
}

func secureOpenedPrivateJSONLFile(path string, file *os.File) error {
	if file == nil {
		return fmt.Errorf("agentx/telemetry: observation file is unavailable")
	}
	openedInfo, err := file.Stat()
	if err != nil {
		return err
	}
	currentInfo, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !openedInfo.Mode().IsRegular() || currentInfo.Mode()&os.ModeSymlink != 0 || !currentInfo.Mode().IsRegular() || !os.SameFile(openedInfo, currentInfo) {
		return fmt.Errorf("agentx/telemetry: observation path identity changed while opening: %s", path)
	}
	if err := file.Chmod(privateObservationFileMode); err != nil {
		return err
	}
	securedInfo, err := file.Stat()
	if err != nil {
		return err
	}
	if securedInfo.Mode().Perm() != privateObservationFileMode.Perm() {
		return fmt.Errorf("agentx/telemetry: observation file mode verification failed: got %#o want %#o", securedInfo.Mode().Perm(), privateObservationFileMode)
	}
	return nil
}
