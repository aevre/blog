package main

import (
	"testing"

	"go.uber.org/fx"
)

func TestApplicationGraph(t *testing.T) {
	if err := fx.ValidateApp(applicationOptions()); err != nil {
		t.Fatal(err)
	}
}
