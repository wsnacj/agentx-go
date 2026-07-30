package promptcontext

import "time"

type Context struct {
	Now       time.Time
	Timezone  string
	SessionID string
	Model     string
}

type BuildInput struct {
	Now       time.Time
	Timezone  string
	SessionID string
	Model     string
}

func Build(input BuildInput) Context {
	now := input.Now
	if now.IsZero() {
		now = time.Now()
	}
	return Context{
		Now:       now,
		Timezone:  input.Timezone,
		SessionID: input.SessionID,
		Model:     input.Model,
	}
}

func (c Context) TimestampText() string {
	if c.Timezone == "" {
		return c.Now.Format(time.RFC3339)
	}
	loc, err := time.LoadLocation(c.Timezone)
	if err != nil {
		return c.Now.Format(time.RFC3339)
	}
	return c.Now.In(loc).Format(time.RFC3339)
}
