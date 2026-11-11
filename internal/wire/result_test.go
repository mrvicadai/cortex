package wire_test

import (
	"strings"
	"testing"

	"github.com/mrvicadai/cortex/internal/wire"
)

func TestWireBudget64K(t *testing.T) {
	if wire.Budget != 65536 {
		t.Fatalf("expected wire.Budget to be 65536 (64 KiB), got %d", wire.Budget)
	}

	// Small payload fits
	small := map[string]string{"message": "hello world"}
	if !wire.Fits(small) {
		t.Errorf("expected small payload to fit")
	}

	// Payload around 25 KiB raw string (~50 KiB wire frame) fits within 64 KiB
	payload25K := map[string]string{"data": strings.Repeat("a", 25000)}
	if !wire.Fits(payload25K) {
		t.Errorf("expected 25 KiB payload to fit within 64 KiB wire budget")
	}

	// Huge payload exceeds budget
	huge := map[string]string{"data": strings.Repeat("x", 40000)}
	if wire.Fits(huge) {
		t.Errorf("expected huge payload (>80 KiB wire frame) to exceed 64 KiB wire budget")
	}
}
