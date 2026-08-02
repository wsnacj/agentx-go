package textin

import (
	"github.com/wsnacj/agentx-go/document/ocr/diff"
	"github.com/wsnacj/agentx-go/document/ocr/model"
)

func summarizeOCRDiff(res diff.DiffResult, previewLimit int) *model.DiffResultSummary {
	preview := collectPreview(res.PageDiffs, previewLimit)
	return &model.DiffResultSummary{
		HasDiff:       res.HasDiff,
		ExtraPages:    len(res.ExtraPages),
		DiffPageCount: countDiffPages(res.PageDiffs),
		MappingScheme: res.MappingScheme,
		Preview:       preview,
	}
}

func summarizeTableDiff(res diff.TableDiffResult, previewLimit int) *model.DiffResultSummary {
	preview := collectPreviewTable(res.PageDiffs, previewLimit)
	return &model.DiffResultSummary{
		HasDiff:       res.HasDiff,
		ExtraPages:    len(res.ExtraPages),
		DiffPageCount: countDiffPagesTable(res.PageDiffs),
		MappingScheme: res.MappingScheme,
		Preview:       preview,
	}
}

func collectPreview(pages []diff.PageDiff, limit int) []string {
	var previews []string
	for _, pd := range pages {
		if len(pd.DiffAreas) == 0 {
			continue
		}
		previews = append(previews, pd.DiffAreas[0].DiffText)
		if limit > 0 && len(previews) >= limit {
			break
		}
	}
	return previews
}

func collectPreviewTable(pages []diff.TablePageDiff, limit int) []string {
	var previews []string
	for _, pd := range pages {
		if len(pd.DiffAreas) == 0 {
			continue
		}
		previews = append(previews, pd.DiffAreas[0].DiffText)
		if limit > 0 && len(previews) >= limit {
			break
		}
	}
	return previews
}

func countDiffPages(pages []diff.PageDiff) int {
	count := 0
	for _, pd := range pages {
		if len(pd.DiffAreas) > 0 {
			count++
		}
	}
	return count
}

func countDiffPagesTable(pages []diff.TablePageDiff) int {
	count := 0
	for _, pd := range pages {
		if len(pd.DiffAreas) > 0 {
			count++
		}
	}
	return count
}
