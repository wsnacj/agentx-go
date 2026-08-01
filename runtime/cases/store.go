package cases

import "context"

type Store interface {
	UpsertCase(ctx context.Context, value Case) error
	GetCase(ctx context.Context, caseID string) (Case, error)
	ListCases(ctx context.Context, filter Filter) ([]Case, error)
}
