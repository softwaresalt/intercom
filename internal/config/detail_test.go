package config

import (
	"errors"
	"testing"

	"github.com/softwaresalt/intercom-go/internal/apperr"
)

func TestOperatorDetailLevelUnmarshalTextAcceptsValidValues(t *testing.T) {
	tests := []struct {
		text string
		want OperatorDetailLevel
	}{
		{"minimal", DetailMinimal},
		{"standard", DetailStandard},
		{"verbose", DetailVerbose},
	}

	for _, tt := range tests {
		var d OperatorDetailLevel
		if err := d.UnmarshalText([]byte(tt.text)); err != nil {
			t.Fatalf("UnmarshalText(%q) returned unexpected error: %v", tt.text, err)
		}
		if d != tt.want {
			t.Errorf("UnmarshalText(%q) = %v, want %v", tt.text, d, tt.want)
		}
	}
}

func TestOperatorDetailLevelUnmarshalTextRejectsUnknownValue(t *testing.T) {
	var d OperatorDetailLevel
	err := d.UnmarshalText([]byte("loud"))
	if err == nil {
		t.Fatal("UnmarshalText(\"loud\") returned nil error, want KindConfig error")
	}

	var appErr *apperr.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("UnmarshalText(\"loud\") error is not *apperr.Error: %v", err)
	}
	if appErr.Kind() != apperr.KindConfig {
		t.Errorf("UnmarshalText(\"loud\") error kind = %v, want KindConfig", appErr.Kind())
	}
}
