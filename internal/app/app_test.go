package app

import "testing"

func TestBanner(t *testing.T) {
	t.Parallel()

	const want = "openbudget: Budget tooling is under construction."

	if got := Banner(); got != want {
		t.Fatalf("Banner() = %q, want %q", got, want)
	}
}
