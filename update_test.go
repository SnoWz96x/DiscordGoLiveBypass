package main

import "testing"

func TestParseVersion(t *testing.T) {
	got := parseVersion(" v1.2.3 ")
	want := []int{1, 2, 3}

	if len(got) != len(want) {
		t.Fatalf("parseVersion = %v, queria %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("parseVersion = %v, queria %v", got, want)
		}
	}

	// Sem numero nao ha comparacao, e a versao nova deixa de ser anunciada em vez de virar
	// alarme falso: e o caso do binario compilado na mao e o do pre-lancamento.
	for _, raw := range []string{"dev", "v2.0.0-beta", ""} {
		if got := parseVersion(raw); got != nil {
			t.Errorf("parseVersion(%q) = %v, queria nil", raw, got)
		}
	}
}

func TestNewerThanRelease(t *testing.T) {
	if !newerThan(parseVersion("v1.2.0"), parseVersion("v1.1.9")) {
		t.Error("v1.2.0 deveria ser mais nova que v1.1.9")
	}
	if newerThan(parseVersion("v1.2.0"), parseVersion("v1.2.0")) {
		t.Error("a mesma versao nao deveria virar aviso de atualizacao")
	}
}
