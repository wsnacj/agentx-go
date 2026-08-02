package pipeline

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/wsnacj/agentx-go/document/pipeline/types"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const parseFingerprintAlgorithm = "docparse-fingerprint-v1"

type parseFingerprintInput struct {
	DocPath        string
	SpecPath       string
	ModelName      string
	PDFParseMode   PDFParseMode
	ExtractionMode DocumentExtractionMode
	MaxChunkChars  int
	PageLimit      int
}

func buildParseFingerprint(input parseFingerprintInput) (*types.ParseFingerprint, error) {
	docHash, err := fileSHA256(input.DocPath)
	if err != nil {
		return nil, fmt.Errorf("hash document: %w", err)
	}
	specHash, specFileCount, err := specTreeSHA256(input.SpecPath)
	if err != nil {
		return nil, fmt.Errorf("hash spec: %w", err)
	}
	modeName := pdfParseModeName(input.PDFParseMode)
	extractionMode := documentExtractionModeName(input.ExtractionMode)
	modelName := strings.TrimSpace(input.ModelName)
	cacheHash := sha256.New()
	writeHashPart(cacheHash, "algorithm", parseFingerprintAlgorithm)
	writeHashPart(cacheHash, "document_sha256", docHash)
	writeHashPart(cacheHash, "spec_sha256", specHash)
	writeHashPart(cacheHash, "model_name", modelName)
	writeHashPart(cacheHash, "pdf_parse_mode", modeName)
	writeHashPart(cacheHash, "extraction_mode", extractionMode)
	writeHashPart(cacheHash, "max_chunk_chars", fmt.Sprint(input.MaxChunkChars))
	writeHashPart(cacheHash, "page_limit", fmt.Sprint(input.PageLimit))
	return &types.ParseFingerprint{
		Algorithm:      parseFingerprintAlgorithm,
		CacheKey:       hex.EncodeToString(cacheHash.Sum(nil)),
		DocumentSHA256: docHash,
		SpecSHA256:     specHash,
		SpecFileCount:  specFileCount,
		ModelName:      modelName,
		PDFParseMode:   modeName,
		ExtractionMode: extractionMode,
		MaxChunkChars:  input.MaxChunkChars,
		PageLimit:      input.PageLimit,
	}, nil
}

func fileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func specTreeSHA256(path string) (string, int, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", 0, err
	}
	if !info.IsDir() {
		hash, err := fileSHA256(path)
		if err != nil {
			return "", 0, err
		}
		return hash, 1, nil
	}
	files := []string{}
	if err := filepath.WalkDir(path, func(current string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		files = append(files, current)
		return nil
	}); err != nil {
		return "", 0, err
	}
	sort.Strings(files)
	hasher := sha256.New()
	for _, filePath := range files {
		rel, err := filepath.Rel(path, filePath)
		if err != nil {
			return "", 0, err
		}
		contentHash, err := fileSHA256(filePath)
		if err != nil {
			return "", 0, err
		}
		writeHashPart(hasher, filepath.ToSlash(rel), contentHash)
	}
	return hex.EncodeToString(hasher.Sum(nil)), len(files), nil
}

func writeHashPart(writer io.Writer, key string, value string) {
	_, _ = io.WriteString(writer, key)
	_, _ = io.WriteString(writer, "\x00")
	_, _ = io.WriteString(writer, value)
	_, _ = io.WriteString(writer, "\x00")
}

func pdfParseModeName(mode PDFParseMode) string {
	switch mode {
	case PDFParseSimple:
		return "simple"
	case PDFParseNormal:
		return "normal"
	case PDFParseForceOCR:
		return "force_ocr"
	default:
		return fmt.Sprintf("unknown_%d", int(mode))
	}
}

func documentExtractionModeName(mode DocumentExtractionMode) string {
	normalized, err := NormalizeDocumentExtractionMode(string(mode))
	if err != nil {
		return string(mode)
	}
	return string(normalized)
}
