package ocrx

import "github.com/wsnacj/agentx-go/document/ocr/model"

type (
	OperationKind = model.OperationKind
	Request       = model.Request
	Response      = model.Response
	Meta          = model.Meta

	OCRPayload    = model.OCRPayload
	TablePayload  = model.TablePayload
	TablePage     = model.TablePage
	TableResponse = model.TableResponse

	StampPayload  = model.StampPayload
	StampPage     = model.StampPage
	StampDetail   = model.StampDetail
	StampResponse = model.StampResponse

	DiffSummary       = model.DiffSummary
	DiffResultSummary = model.DiffResultSummary

	Coordinate = model.Coordinate
	TextBox    = model.TextBox
)

const (
	OperationKindAny   = model.OperationKindAny
	OperationKindOCR   = model.OperationKindOCR
	OperationKindTable = model.OperationKindTable
	OperationKindStamp = model.OperationKindStamp
)
