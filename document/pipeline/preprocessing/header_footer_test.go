package preprocessing

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProgrammaticCleanupIgnoresCoverPage(t *testing.T) {
	pages := []string{
		"封面\n正文第一页",
		"ACME Confidential\nPage 2 of 3\n正文第二页",
		"ACME Confidential\nPage 3 of 3\n正文第三页",
	}
	cleaned, err := programmaticCleanup(pages, 2, 0)
	require.NoError(t, err)
	require.Len(t, cleaned, len(pages))
	assert.Equal(t, "封面\n正文第一页", cleaned[0])
	assert.Equal(t, "正文第二页", cleaned[1])
	assert.Equal(t, "正文第三页", cleaned[2])
}

func TestProgrammaticCleanupAvoidsSectionTitleFalsePositive(t *testing.T) {
	pages := []string{
		"Chapter 1 Overview\n正文 A",
		"Chapter 2 Design\n正文 B",
		"Chapter 3 Result\n正文 C",
	}
	cleaned, err := programmaticCleanup(pages, 2, 0)
	require.NoError(t, err)
	assert.Equal(t, pages, cleaned)
}

func TestLLMCleanupUsesConsensusPatterns(t *testing.T) {
	request := func(context.Context, string, string, []string) (string, error) {
		return "HEADER_LINES:2\nFOOTER_LINES:0", nil
	}
	pages := []string{
		"ACME Confidential\nPage 1 of 2\n正文一",
		"ACME Confidential\nPage 2 of 2\n正文二",
	}
	cleaned, err := llmCleanup(context.Background(), pages, "fake-model", 2, request)
	require.NoError(t, err)
	assert.Equal(t, "正文一", cleaned[0])
	assert.Equal(t, "正文二", cleaned[1])
}

func TestLLMCleanupSkipsWhenConsensusMissing(t *testing.T) {
	request := func(context.Context, string, string, []string) (string, error) {
		return "HEADER_LINES:2", nil
	}
	pages := []string{"Unique Header 1\n正文一", "Unique Header 2\n正文二"}
	cleaned, err := llmCleanup(context.Background(), pages, "fake-model", 2, request)
	require.NoError(t, err)
	assert.Equal(t, pages, cleaned)
}

func TestLLMCleanupPropagatesError(t *testing.T) {
	request := func(context.Context, string, string, []string) (string, error) {
		return "", errors.New("boom")
	}
	_, err := llmCleanup(context.Background(), []string{"a"}, "fake-model", 1, request)
	assert.Error(t, err)
}
