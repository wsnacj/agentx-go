package channel

import (
	"context"
	"fmt"
	"strings"
)

type ChunkingSender struct {
	Base  TextSender
	Limit int
}

func (s ChunkingSender) SendText(ctx context.Context, target TextTarget, text string) error {
	if s.Base == nil {
		return nil
	}
	for _, chunk := range SplitText(text, s.Limit) {
		if err := s.Base.SendText(ctx, target, chunk); err != nil {
			return err
		}
	}
	return nil
}

func (s ChunkingSender) ReplyText(ctx context.Context, target TextTarget, text string) error {
	if s.Base == nil {
		return nil
	}
	replySender, ok := s.Base.(ReplySender)
	if !ok {
		return s.SendText(ctx, target, text)
	}
	for _, chunk := range SplitText(text, s.Limit) {
		if err := replySender.ReplyText(ctx, target, chunk); err != nil {
			return err
		}
	}
	return nil
}

func (s ChunkingSender) EditText(ctx context.Context, target TextTarget, text string) error {
	if s.Base == nil {
		return nil
	}
	editSender, ok := s.Base.(EditSender)
	if !ok {
		return fmt.Errorf("channel edit is unsupported by the wrapped sender")
	}
	return editSender.EditText(ctx, target, strings.TrimSpace(text))
}

func (s ChunkingSender) DeleteMessage(ctx context.Context, target TextTarget) error {
	if s.Base == nil {
		return nil
	}
	deleteSender, ok := s.Base.(DeleteSender)
	if !ok {
		return nil
	}
	return deleteSender.DeleteMessage(ctx, target)
}

func (s ChunkingSender) ReactMessage(ctx context.Context, target TextTarget, emoji string, remove bool, reactionID string) error {
	if s.Base == nil {
		return nil
	}
	reactSender, ok := s.Base.(ReactSender)
	if !ok {
		return fmt.Errorf("channel react is unsupported by the wrapped sender")
	}
	return reactSender.ReactMessage(ctx, target, strings.TrimSpace(emoji), remove, strings.TrimSpace(reactionID))
}

func (s ChunkingSender) ForwardMessage(ctx context.Context, target TextTarget, sourceMessageID string) error {
	if s.Base == nil {
		return nil
	}
	forwardSender, ok := s.Base.(ForwardSender)
	if !ok {
		return fmt.Errorf("channel forward is unsupported by the wrapped sender")
	}
	return forwardSender.ForwardMessage(ctx, target, strings.TrimSpace(sourceMessageID))
}

func SplitText(text string, limit int) []string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil
	}
	if limit <= 0 || runeLen(trimmed) <= limit {
		return []string{trimmed}
	}
	paragraphs := strings.Split(trimmed, "\n")
	chunks := make([]string, 0, len(paragraphs))
	current := ""
	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}
		if runeLen(paragraph) > limit {
			if strings.TrimSpace(current) != "" {
				chunks = append(chunks, strings.TrimSpace(current))
				current = ""
			}
			chunks = append(chunks, splitLongSegment(paragraph, limit)...)
			continue
		}
		candidate := paragraph
		if current != "" {
			candidate = current + "\n" + paragraph
		}
		if runeLen(candidate) <= limit {
			current = candidate
			continue
		}
		if strings.TrimSpace(current) != "" {
			chunks = append(chunks, strings.TrimSpace(current))
		}
		current = paragraph
	}
	if strings.TrimSpace(current) != "" {
		chunks = append(chunks, strings.TrimSpace(current))
	}
	if len(chunks) == 0 {
		return []string{trimmed}
	}
	return chunks
}

func splitLongSegment(text string, limit int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return splitByRune(text, limit)
	}
	chunks := make([]string, 0, len(words))
	current := ""
	for _, word := range words {
		if runeLen(word) > limit {
			if current != "" {
				chunks = append(chunks, current)
				current = ""
			}
			chunks = append(chunks, splitByRune(word, limit)...)
			continue
		}
		candidate := word
		if current != "" {
			candidate = current + " " + word
		}
		if runeLen(candidate) <= limit {
			current = candidate
			continue
		}
		chunks = append(chunks, current)
		current = word
	}
	if current != "" {
		chunks = append(chunks, current)
	}
	return chunks
}

func splitByRune(text string, limit int) []string {
	runes := []rune(strings.TrimSpace(text))
	if len(runes) == 0 {
		return nil
	}
	if limit <= 0 || len(runes) <= limit {
		return []string{string(runes)}
	}
	out := make([]string, 0, (len(runes)+limit-1)/limit)
	for start := 0; start < len(runes); start += limit {
		end := start + limit
		if end > len(runes) {
			end = len(runes)
		}
		out = append(out, strings.TrimSpace(string(runes[start:end])))
	}
	return out
}

func runeLen(text string) int {
	return len([]rune(text))
}
