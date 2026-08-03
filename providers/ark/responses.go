package ark

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/wsnacj/agentx-go/providers/ark/types"
)

// CreateResponse sends one raw Ark Responses request without model routing or fallback policy.
func (c *Client) CreateResponse(ctx context.Context, request types.ResponseRequest) (*types.Response, error) {
	if c == nil {
		return nil, fmt.Errorf("ark client: not initialized")
	}
	if err := types.ValidateTools(request.Tools); err != nil {
		return nil, err
	}
	var response types.Response
	if err := c.DoJSON(ctx, http.MethodPost, "/responses", request, &response); err != nil {
		return nil, err
	}
	response.RequestModel = request.Model
	return &response, nil
}

// GetResponse retrieves one response by ID.
func (c *Client) GetResponse(ctx context.Context, id string) (*types.Response, error) {
	var response types.Response
	if err := c.DoJSON(ctx, http.MethodGet, "/responses/"+id, nil, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// DeleteResponse deletes one response by ID.
func (c *Client) DeleteResponse(ctx context.Context, id string) error {
	return c.DoJSON(ctx, http.MethodDelete, "/responses/"+id, nil, nil)
}

// StreamResponse opens and decodes one raw Ark Responses SSE stream.
func (c *Client) StreamResponse(ctx context.Context, request types.ResponseRequest) (*types.StreamResult, error) {
	if c == nil {
		return nil, fmt.Errorf("ark client: not initialized")
	}
	if err := types.ValidateTools(request.Tools); err != nil {
		return nil, err
	}
	stream := request
	streaming := true
	stream.Stream = &streaming
	body, cancel, err := c.DoStream(ctx, "/responses", stream)
	if err != nil {
		return nil, err
	}
	events := make(chan types.Event)
	errs := make(chan error, 1)
	go func() {
		defer close(events)
		defer close(errs)
		defer body.Close()
		defer cancel()
		if err := decodeSSE(body, events); err != nil && err != io.EOF {
			errs <- err
		}
	}()
	return &types.StreamResult{Ch: events, Err: errs, Cancel: func() { cancel(); _ = body.Close() }}, nil
}

// ContinueWithToolOutputs builds and sends the next response request.
func (c *Client) ContinueWithToolOutputs(ctx context.Context, previous *types.Response, results []FunctionResult, options ToolFollowupOptions) (*types.Response, error) {
	request, err := BuildToolOutputRequest(previous, results, options)
	if err != nil {
		return nil, err
	}
	return c.CreateResponse(ctx, request)
}

// CreateImageGeneration sends one non-streaming Images request.
func (c *Client) CreateImageGeneration(ctx context.Context, request types.ImageGenerationRequest) (*types.ImageGenerationResponse, error) {
	if request.Stream != nil && *request.Stream {
		return nil, fmt.Errorf("ark image generation: stream=true is not supported in CreateImageGeneration, use StreamImageGeneration")
	}
	var response types.ImageGenerationResponse
	if err := c.DoJSON(ctx, http.MethodPost, "/images/generations", request, &response); err != nil {
		return nil, err
	}
	response.RequestModel = request.Model
	return &response, nil
}

// StreamImageGeneration opens and decodes one Images SSE stream.
func (c *Client) StreamImageGeneration(ctx context.Context, request types.ImageGenerationRequest) (*types.ImageGenerationStreamResult, error) {
	stream := request
	streaming := true
	stream.Stream = &streaming
	body, cancel, err := c.DoStream(ctx, "/images/generations", stream)
	if err != nil {
		return nil, err
	}
	events := make(chan types.ImageGenerationEvent)
	errs := make(chan error, 1)
	go func() {
		defer close(events)
		defer close(errs)
		defer body.Close()
		defer cancel()
		if err := decodeImageGenerationSSE(body, events); err != nil && err != io.EOF {
			errs <- err
		}
	}()
	return &types.ImageGenerationStreamResult{Ch: events, Err: errs, Cancel: func() { cancel(); _ = body.Close() }}, nil
}

func decodeSSE(reader io.Reader, out chan<- types.Event) error {
	return decodeDataSSE(reader, func(data []byte) error {
		event, err := types.DecodeEvent(data)
		if err == nil {
			out <- event
		}
		return err
	})
}

func decodeImageGenerationSSE(reader io.Reader, out chan<- types.ImageGenerationEvent) error {
	return decodeDataSSE(reader, func(data []byte) error {
		event, err := types.DecodeImageGenerationEvent(data)
		if err == nil {
			out <- event
		}
		return err
	})
}

func decodeDataSSE(source io.Reader, emit func([]byte) error) error {
	reader := bufio.NewReader(source)
	var data strings.Builder
	flush := func() error {
		if data.Len() == 0 {
			return nil
		}
		payload := strings.TrimSpace(data.String())
		data.Reset()
		if payload == "" {
			return nil
		}
		if payload == "[DONE]" {
			return io.EOF
		}
		return emit([]byte(payload))
	}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if flushErr := flush(); flushErr != nil && flushErr != io.EOF {
					return flushErr
				}
				return nil
			}
			return err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			if err := flush(); err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
			continue
		}
		if strings.HasPrefix(line, "data:") {
			if data.Len() > 0 {
				data.WriteByte('\n')
			}
			data.WriteString(strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}
}
