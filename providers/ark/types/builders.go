package types

// NewMessage builds a role-based input message.
func NewMessage(role string, content InputContent) InputItem {
	return InputItem{
		Role:    role,
		Content: content,
	}
}

// NewInputTextItem creates an input text content item.
func NewInputTextItem(text string) ContentItem {
	return ContentItem{
		Type: "input_text",
		Text: text,
	}
}

// NewInputImageURLItem creates an input image content item from URL/base64.
func NewInputImageURLItem(url string, detail string) ContentItem {
	return ContentItem{
		Type:     "input_image",
		ImageURL: &ImageURL{Raw: url},
		Detail:   detail,
	}
}

// NewInputImageFileItem creates an input image content item from file ID.
func NewInputImageFileItem(fileID string, detail string) ContentItem {
	return ContentItem{
		Type:   "input_image",
		FileID: fileID,
		Detail: detail,
	}
}

// NewInputImageURLItemWithLimit creates an input image content item with pixel limits.
func NewInputImageURLItemWithLimit(url string, detail string, limit *ImagePixelLimit) ContentItem {
	return ContentItem{
		Type:            "input_image",
		ImageURL:        &ImageURL{Raw: url},
		Detail:          detail,
		ImagePixelLimit: limit,
	}
}

// NewInputImageFileItemWithLimit creates an input image content item with pixel limits.
func NewInputImageFileItemWithLimit(fileID string, detail string, limit *ImagePixelLimit) ContentItem {
	return ContentItem{
		Type:            "input_image",
		FileID:          fileID,
		Detail:          detail,
		ImagePixelLimit: limit,
	}
}

// NewInputFileItem creates an input file content item from file ID.
func NewInputFileItem(fileID string) ContentItem {
	return ContentItem{
		Type:   "input_file",
		FileID: fileID,
	}
}

// NewInputVideoFileItem creates an input video content item from file ID.
func NewInputVideoFileItem(fileID string) ContentItem {
	return ContentItem{
		Type:   "input_video",
		FileID: fileID,
	}
}

// NewImagePixelLimit builds pixel limit settings.
func NewImagePixelLimit(minPixels int, maxPixels int) *ImagePixelLimit {
	if minPixels == 0 && maxPixels == 0 {
		return nil
	}
	return &ImagePixelLimit{MinPixels: minPixels, MaxPixels: maxPixels}
}
