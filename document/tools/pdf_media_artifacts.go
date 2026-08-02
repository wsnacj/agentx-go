package tools

import (
	"context"
	"fmt"
	"strings"

	agentxmedia "github.com/wsnacj/agentx-go/runtime/mediaartifact"
)

func persistPDFRenderedPageArtifacts(ctx context.Context, root string, sourcePath string, rendered []pdfRenderedPage) ([]agentxmedia.Descriptor, error) {
	if strings.TrimSpace(root) == "" || len(rendered) == 0 {
		return nil, nil
	}
	host := pdfHostFromContext(ctx)
	if host.PublishRendered == nil {
		return nil, fmt.Errorf("pdf rendered artifact publisher is not configured")
	}
	return host.PublishRendered(ctx, root, sourcePath, append([]pdfRenderedPage(nil), rendered...))
}
