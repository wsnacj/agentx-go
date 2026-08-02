package ocrx

import (
	"context"
	"fmt"
	"sync"

	"github.com/wsnacj/agentx-go/document/ocr/cache"
	"github.com/wsnacj/agentx-go/document/ocr/config"
	"github.com/wsnacj/agentx-go/document/ocr/model"
	"github.com/wsnacj/agentx-go/document/ocr/pipeline"
	"github.com/wsnacj/agentx-go/document/ocr/processor"
	baiduproc "github.com/wsnacj/agentx-go/document/ocr/processor/baidu"
	textinproc "github.com/wsnacj/agentx-go/document/ocr/processor/textin"
	volcproc "github.com/wsnacj/agentx-go/document/ocr/processor/volcengine"
	"github.com/wsnacj/agentx-go/document/ocr/provider"
	"github.com/wsnacj/agentx-go/document/ocr/splitter"
	"github.com/wsnacj/agentx-go/document/ocr/worker"
)

// Service exposes the main OCR, table, and stamp recognition entry points.
// It acts as a façade coordinating the lower-level pipelines.
type Service struct {
	mu        sync.RWMutex
	cfg       config.ServiceConfig
	pipelines map[model.OperationKind]*pipeline.Manager
}

// NewService constructs a Service from the provided configuration bundle.
func NewService(cfg config.ServiceConfig, deps Dependencies) (*Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid ocrx configuration: %w", err)
	}
	deps = deps.withDefaults()

	pipelines := make(map[model.OperationKind]*pipeline.Manager)
	for key, pcfg := range cfg.Pipelines {
		kind := model.OperationKind(key)

		if validator, ok := deps.ProviderConfigValidators[pcfg.Provider.Kind]; ok {
			if err := validator(&pcfg.Provider); err != nil {
				return nil, fmt.Errorf("validate provider config for %s: %w", kind, err)
			}
		}

		provFactory, ok := deps.ProviderFactories[pcfg.Provider.Kind]
		if !ok {
			return nil, fmt.Errorf("no provider factory registered for kind %s", pcfg.Provider.Kind)
		}

		prov, err := provFactory(pcfg.Provider)
		if err != nil {
			return nil, fmt.Errorf("instantiate provider for %s: %w", kind, err)
		}

		procFactory, err := deps.ProcessorFactories.Lookup(pcfg.Provider.Kind)
		if err != nil {
			return nil, fmt.Errorf("processor factory for %s: %w", pcfg.Provider.Kind, err)
		}
		providerProcessor, err := procFactory(pcfg.Provider)
		if err != nil {
			return nil, fmt.Errorf("instantiate processor for %s: %w", kind, err)
		}
		opProcessor, err := providerProcessor.For(kind)
		if err != nil {
			return nil, fmt.Errorf("processor not available for %s: %w", kind, err)
		}

		splitterFactory, ok := deps.SplitterFactories[pcfg.Splitter.Kind]
		if !ok {
			return nil, fmt.Errorf("no splitter factory registered for kind %s", pcfg.Splitter.Kind)
		}
		spli, err := splitterFactory(pcfg.Splitter)
		if err != nil {
			return nil, fmt.Errorf("instantiate splitter for %s: %w", kind, err)
		}

		cacheStore, err := deps.CacheBuilder(pcfg.Cache)
		if err != nil {
			return nil, fmt.Errorf("instantiate cache for %s: %w", kind, err)
		}

		wp := worker.NewPool(pcfg.Worker)

		manager, err := pipeline.NewManager(kind, pcfg, prov, spli, cacheStore, wp, opProcessor)
		if err != nil {
			return nil, fmt.Errorf("new pipeline manager for %s: %w", kind, err)
		}
		pipelines[kind] = manager
		cfg.Pipelines[key] = pcfg
	}

	return &Service{
		cfg:       cfg,
		pipelines: pipelines,
	}, nil
}

// RecognizeOCR runs the standard OCR pipeline.
func (s *Service) RecognizeOCR(ctx context.Context, req model.Request) (model.Response, error) {
	return s.run(ctx, model.OperationKindOCR, req)
}

// RecognizeTable runs the table OCR pipeline.
func (s *Service) RecognizeTable(ctx context.Context, req model.Request) (model.TableResponse, error) {
	resp, err := s.run(ctx, model.OperationKindTable, req)
	if err != nil {
		return model.TableResponse{}, err
	}
	tableResp, ok := resp.Payload.(model.TablePayload)
	if !ok {
		return model.TableResponse{}, fmt.Errorf("unexpected table payload type %T", resp.Payload)
	}
	return model.TableResponse{
		Meta:    resp.Meta,
		Payload: tableResp,
		Diff:    resp.Diff,
	}, nil
}

// RecognizeStamp runs the stamp detection pipeline.
func (s *Service) RecognizeStamp(ctx context.Context, req model.Request) (model.StampResponse, error) {
	resp, err := s.run(ctx, model.OperationKindStamp, req)
	if err != nil {
		return model.StampResponse{}, err
	}
	stampPayload, ok := resp.Payload.(model.StampPayload)
	if !ok {
		return model.StampResponse{}, fmt.Errorf("unexpected stamp payload type %T", resp.Payload)
	}
	return model.StampResponse{
		Meta:    resp.Meta,
		Payload: stampPayload,
	}, nil
}

func (s *Service) run(ctx context.Context, kind model.OperationKind, req model.Request) (model.Response, error) {
	s.mu.RLock()
	manager, ok := s.pipelines[kind]
	s.mu.RUnlock()
	if !ok {
		return model.Response{}, fmt.Errorf("pipeline not configured for %s", kind)
	}
	return manager.Run(ctx, req)
}

// Dependencies groups optional factories used to build the Service instance.
type Dependencies struct {
	ProviderFactories        map[string]provider.Factory
	SplitterFactories        map[string]splitter.Factory
	CacheBuilder             cache.Builder
	ProcessorFactories       processor.Registry
	ProviderConfigValidators map[string]provider.ConfigValidator
}

func (d Dependencies) withDefaults() Dependencies {
	defaultProviders := provider.DefaultFactories()
	if d.ProviderFactories == nil {
		d.ProviderFactories = defaultProviders
	} else {
		for k, v := range defaultProviders {
			if _, ok := d.ProviderFactories[k]; !ok {
				d.ProviderFactories[k] = v
			}
		}
	}
	if d.SplitterFactories == nil {
		d.SplitterFactories = splitter.DefaultFactories()
	}
	if d.CacheBuilder == nil {
		d.CacheBuilder = cache.DefaultBuilder()
	}
	defaultProcessors := processor.Registry{
		"textin": func(cfg config.ProviderConfig) (processor.ProviderProcessor, error) {
			return textinproc.NewProcessor(cfg)
		},
		"volcengine": func(cfg config.ProviderConfig) (processor.ProviderProcessor, error) {
			return volcproc.NewProcessor(cfg)
		},
		"baidu": func(cfg config.ProviderConfig) (processor.ProviderProcessor, error) {
			return baiduproc.NewProcessor(cfg)
		},
	}
	if d.ProcessorFactories == nil {
		d.ProcessorFactories = defaultProcessors
	} else {
		for k, v := range defaultProcessors {
			if _, ok := d.ProcessorFactories[k]; !ok {
				d.ProcessorFactories[k] = v
			}
		}
	}
	defaultValidators := provider.DefaultConfigValidators()
	if d.ProviderConfigValidators == nil {
		d.ProviderConfigValidators = defaultValidators
	} else {
		for k, v := range defaultValidators {
			if _, ok := d.ProviderConfigValidators[k]; !ok {
				d.ProviderConfigValidators[k] = v
			}
		}
	}
	return d
}
