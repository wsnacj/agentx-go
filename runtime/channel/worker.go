package channel

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"
)

type AccessDecision struct {
	Blocked bool
	Reply   string
}

type DedupeReservationBuilder func(Message) []DedupeReservation
type AccessEvaluator func(Message) (AccessDecision, error)
type EventLogger func(format string, args ...any)

type InboundProcessor struct {
	Runner  TurnRunner
	Sender  TextSender
	Deduper *Deduper
	// Ingress is the explicit owner for asynchronous processing.
	Ingress           *IngressRuntime
	Timeout           time.Duration
	Logger            EventLogger
	EvaluateAccess    AccessEvaluator
	BuildReservations DedupeReservationBuilder
}

func NewWriterLogger(w io.Writer) EventLogger {
	if w == nil {
		return nil
	}
	return func(format string, args ...any) {
		_, _ = fmt.Fprintf(w, format, args...)
	}
}

func (p InboundProcessor) Process(ctx context.Context, message Message) error {
	reservations, ok := p.beginDedupe(message)
	if !ok {
		return nil
	}
	return p.processReserved(ctx, message, reservations)
}

// ProcessAsync submits to Ingress and reports whether the message was accepted.
func (p InboundProcessor) ProcessAsync(ctx context.Context, message Message) bool {
	return p.SubmitAsync(ctx, message).Accepted
}

// SubmitAsync submits to Ingress and returns a structured acceptance reason.
func (p InboundProcessor) SubmitAsync(ctx context.Context, message Message) IngressSubmitResult {
	result := IngressSubmitResult{Reason: IngressSubmitRuntimeUnavailable}
	if p.Ingress != nil {
		result = p.Ingress.Submit(ctx, p, message)
	}
	if !result.Accepted && p.Logger != nil {
		if p.Ingress != nil {
			p.Ingress.log(p.Logger, "channel event rejected platform=%s account=%s session=%s chat=%s msg=%s reason=%s\n", message.Platform, message.AccountID, message.SessionID, message.ChatID, message.MessageID, result.Reason)
		} else {
			p.Logger("channel event rejected platform=%s account=%s session=%s chat=%s msg=%s reason=%s\n", message.Platform, message.AccountID, message.SessionID, message.ChatID, message.MessageID, result.Reason)
		}
	}
	return result
}

func (p InboundProcessor) processReserved(ctx context.Context, message Message, reservations []DedupeReservation) error {
	timeout := p.Timeout
	if timeout <= 0 {
		timeout = 90 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if p.EvaluateAccess != nil {
		decision, err := p.EvaluateAccess(message)
		if err != nil {
			p.releaseDedupe(reservations)
			return err
		}
		if decision.Blocked {
			if strings.TrimSpace(decision.Reply) != "" && p.Sender != nil {
				if err := p.Sender.SendText(runCtx, TextTarget{
					AccountID: message.AccountID,
					ChatID:    message.ChatID,
					ThreadID:  message.ThreadID,
					MessageID: message.MessageID,
				}, decision.Reply); err != nil {
					p.releaseDedupe(reservations)
					return fmt.Errorf("send gated reply failed: %w", err)
				}
			}
			if cause := context.Cause(runCtx); cause != nil {
				p.releaseDedupe(reservations)
				return fmt.Errorf("gated reply canceled: %w", cause)
			}
			p.completeDedupe(reservations)
			if p.Logger != nil {
				p.Logger("channel event gated platform=%s account=%s session=%s chat=%s msg=%s\n", message.Platform, message.AccountID, message.SessionID, message.ChatID, message.MessageID)
			}
			return nil
		}
	}
	reply, err := p.Runner.RunTurn(runCtx, message)
	if err != nil {
		p.releaseDedupe(reservations)
		return fmt.Errorf("agent run failed: %w", err)
	}
	if cause := context.Cause(runCtx); cause != nil {
		p.releaseDedupe(reservations)
		return fmt.Errorf("agent run canceled: %w", cause)
	}
	if strings.TrimSpace(reply) != "" && p.Sender != nil {
		if err := p.Sender.SendText(runCtx, TextTarget{
			AccountID: message.AccountID,
			ChatID:    message.ChatID,
			ThreadID:  message.ThreadID,
			MessageID: message.MessageID,
		}, reply); err != nil {
			p.releaseDedupe(reservations)
			return fmt.Errorf("send reply failed: %w", err)
		}
	}
	if cause := context.Cause(runCtx); cause != nil {
		p.releaseDedupe(reservations)
		return fmt.Errorf("send reply canceled: %w", cause)
	}
	p.completeDedupe(reservations)
	if p.Logger != nil {
		p.Logger("channel event handled platform=%s account=%s session=%s chat=%s msg=%s\n", message.Platform, message.AccountID, message.SessionID, message.ChatID, message.MessageID)
	}
	return nil
}

func (p InboundProcessor) beginDedupe(message Message) ([]DedupeReservation, bool) {
	if p.Deduper == nil {
		return nil, true
	}
	build := p.BuildReservations
	if build == nil {
		return nil, true
	}
	reservations := build(message)
	if len(reservations) == 0 {
		return nil, true
	}
	acquired := make([]DedupeReservation, 0, len(reservations))
	for _, item := range reservations {
		if !p.Deduper.BeginFor(item.Key, item.TTL) {
			for _, prior := range acquired {
				p.Deduper.Forget(prior.Key)
			}
			return nil, false
		}
		acquired = append(acquired, item)
	}
	return acquired, true
}

func (p InboundProcessor) completeDedupe(reservations []DedupeReservation) {
	if p.Deduper == nil {
		return
	}
	for _, item := range reservations {
		p.Deduper.CompleteFor(item.Key, item.TTL)
	}
}

func (p InboundProcessor) releaseDedupe(reservations []DedupeReservation) {
	if p.Deduper == nil {
		return
	}
	for _, item := range reservations {
		p.Deduper.Forget(item.Key)
	}
}
