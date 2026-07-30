package promptcontext

import (
	"reflect"
	"testing"
	"time"
)

var (
	_ func(BuildInput) Context = Build
	_ func(Context) string     = Context.TimestampText
)

func TestContextAndBuildInputReflectContract(t *testing.T) {
	want := []fieldContract{
		{name: "Now", typ: reflect.TypeFor[time.Time]()},
		{name: "Timezone", typ: reflect.TypeFor[string]()},
		{name: "SessionID", typ: reflect.TypeFor[string]()},
		{name: "Model", typ: reflect.TypeFor[string]()},
	}
	assertFields(t, reflect.TypeFor[Context](), want)
	assertFields(t, reflect.TypeFor[BuildInput](), want)
}

func TestBuildPreservesTextWithoutNormalization(t *testing.T) {
	now := time.Date(2026, 7, 30, 0, 0, 0, 123, time.FixedZone("input", 3*60*60))
	got := Build(BuildInput{
		Now:       now,
		Timezone:  " UTC ",
		SessionID: " session ",
		Model:     " model ",
	})
	want := Context{
		Now:       now,
		Timezone:  " UTC ",
		SessionID: " session ",
		Model:     " model ",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Build() = %#v, want %#v", got, want)
	}
}

func TestBuildZeroNowUsesCurrentClock(t *testing.T) {
	before := time.Now()
	got := Build(BuildInput{SessionID: "session"})
	after := time.Now()
	if got.Now.Before(before) || got.Now.After(after) {
		t.Fatalf("default Now = %v, want within [%v, %v]", got.Now, before, after)
	}
	if got.SessionID != "session" {
		t.Fatalf("SessionID = %q", got.SessionID)
	}
}

func TestTimestampTextZeroAndFixedZoneContract(t *testing.T) {
	if got, want := (Context{}).TimestampText(), (time.Time{}).Format(time.RFC3339); got != want {
		t.Fatalf("zero TimestampText() = %q, want %q", got, want)
	}

	now := time.Date(2026, 7, 30, 1, 2, 3, 999, time.UTC)
	got := (Context{Now: now, Timezone: "Asia/Shanghai"}).TimestampText()
	if got != "2026-07-30T09:02:03+08:00" {
		t.Fatalf("Shanghai TimestampText() = %q", got)
	}
	if got := (Context{Now: now, Timezone: "invalid/timezone"}).TimestampText(); got != now.Format(time.RFC3339) {
		t.Fatalf("invalid timezone TimestampText() = %q", got)
	}
}

type fieldContract struct {
	name string
	typ  reflect.Type
}

func assertFields(t *testing.T, typ reflect.Type, want []fieldContract) {
	t.Helper()
	if typ.NumField() != len(want) {
		t.Fatalf("%s field count = %d, want %d", typ, typ.NumField(), len(want))
	}
	for i, expected := range want {
		field := typ.Field(i)
		if field.Name != expected.name || field.Type != expected.typ {
			t.Fatalf(
				"%s field[%d] = %s %s, want %s %s",
				typ,
				i,
				field.Name,
				field.Type,
				expected.name,
				expected.typ,
			)
		}
	}
}
