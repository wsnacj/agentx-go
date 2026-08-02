package tools

import (
	"strings"
	"testing"
)

func TestBuildPDFUnifiedStructureItems_SplitsRepeatedFurnitureFromBody(t *testing.T) {
	pageTexts := []PDFPageText{
		{
			Page: 1,
			Text: "Confidential Report\nMaster Agreement body terms\nFooter Shared",
		},
		{
			Page: 2,
			Text: "Confidential Report\nAppendix schedule details\nFooter Shared",
		},
	}
	pageMap := buildPDFAnalyzePageMap(pageTexts, 240)
	items := buildPDFUnifiedStructureItems(pageTexts, pageMap, 2, pdfMediaProfile{})
	if len(items) == 0 {
		t.Fatalf("expected structure items, got none")
	}

	var repeatedHeaderItem *pdfUnifiedStructureItem
	var repeatedFooterItem *pdfUnifiedStructureItem
	var bodyItem *pdfUnifiedStructureItem
	for idx := range items {
		item := &items[idx]
		if item.Page == 1 && item.ContentLayer == pdfUnifiedContentLayerFurniture {
			if strings.Contains(item.Excerpt, "Confidential Report") {
				repeatedHeaderItem = item
			}
			if strings.Contains(item.Excerpt, "Footer Shared") {
				repeatedFooterItem = item
			}
		}
		if item.Page == 1 && item.ContentLayer == pdfUnifiedContentLayerBody {
			bodyItem = item
		}
	}
	if repeatedHeaderItem == nil || repeatedFooterItem == nil {
		t.Fatalf("expected repeated header/footer furniture items on page 1, got %#v", items)
	}
	if bodyItem == nil {
		t.Fatalf("expected body item on page 1, got %#v", items)
	}
	if strings.Contains(bodyItem.Excerpt, "Confidential Report") || strings.Contains(bodyItem.Excerpt, "Footer Shared") {
		t.Fatalf("expected body excerpt without repeated furniture, got %#v", bodyItem)
	}
}

func TestBuildPDFUnifiedStructureItems_DetectsMixedDocumentBlockKinds(t *testing.T) {
	pageTexts := []PDFPageText{
		{
			Page: 1,
			Text: "顺丰速运\n寄件人 王霖\n收件人 李涵\n到付23.00元",
		},
		{
			Page: 2,
			Text: "往来询证函\n本公司沈阳透平机械股份有限公司 与 贵州骐信实业有限公司\n截至2024年12月31日 预收贵公司39,797,000.00元\n2025年3月14日",
		},
		{
			Page: 3,
			Text: "盖章\n授权代表\n签章",
		},
	}
	pageMap := buildPDFAnalyzePageMap(pageTexts, 240)
	items := buildPDFUnifiedStructureItems(pageTexts, pageMap, 3, pdfMediaProfile{})
	if len(items) == 0 {
		t.Fatalf("expected structure items, got none")
	}

	bodyByPage := map[int]pdfUnifiedStructureItem{}
	for _, item := range items {
		if item.ContentLayer == pdfUnifiedContentLayerBody {
			bodyByPage[item.Page] = item
		}
	}
	if got := bodyByPage[1]; got.BlockKind != pdfUnifiedStructureBlockKeyValue || got.Role != pdfUnifiedSegmentLogisticsDoc {
		t.Fatalf("expected page 1 key_value/logistics body, got %#v", got)
	}
	if got := bodyByPage[2]; got.Role != pdfUnifiedSegmentBusinessDoc {
		t.Fatalf("expected page 2 business body, got %#v", got)
	}
	if got := bodyByPage[3]; got.BlockKind != pdfUnifiedStructureBlockSignature || got.Role != pdfUnifiedSegmentSignatureStamp {
		t.Fatalf("expected page 3 signature body, got %#v", got)
	}

	document := pdfUnifiedDocumentArtifacts{
		PageMap:         pageMap,
		TextResult:      PDFTextResult{Pages: pageTexts},
		StructureItems:  items,
		Metadata:        PDFMetadataResult{PageCount: 3},
		DocumentProfile: pdfDocumentProfile{},
		MediaProfile:    pdfMediaProfile{},
	}
	segments := buildPDFUnifiedSegments(document)
	if len(segments) != 3 {
		t.Fatalf("expected three role segments, got %#v", segments)
	}
	if segments[0].Kind != pdfUnifiedSegmentLogisticsDoc || segments[1].Kind != pdfUnifiedSegmentBusinessDoc || segments[2].Kind != pdfUnifiedSegmentSignatureStamp {
		t.Fatalf("unexpected segment order from structure items: %#v", segments)
	}
}
