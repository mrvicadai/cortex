package wire_test

import (
	"strings"
	"testing"

	"github.com/mrvicadai/cortex/internal/wire"
)

func TestWireBudget96K(t *testing.T) {
	if wire.Budget != 98304 {
		t.Fatalf("expected wire.Budget to be 98304 (96 KiB), got %d", wire.Budget)
	}

	// Small payload fits
	small := map[string]string{"message": "hello world"}
	if !wire.Fits(small) {
		t.Errorf("expected small payload to fit")
	}

	// Payload around 40 KiB raw string (~80 KiB wire frame) fits within 96 KiB
	payload40K := map[string]string{"data": strings.Repeat("a", 40000)}
	if !wire.Fits(payload40K) {
		t.Errorf("expected 40 KiB payload to fit within 96 KiB wire budget")
	}

	// Huge payload (>120 KiB wire frame) exceeds budget
	huge := map[string]string{"data": strings.Repeat("x", 65000)}
	if wire.Fits(huge) {
		t.Errorf("expected huge payload (>130 KiB wire frame) to exceed 96 KiB wire budget")
	}
}
