package main

import (
	"strings"
	"testing"
)

func TestParseRoute(t *testing.T) {
	for _, raw := range []string{"all", " PAC-Gateway ", "pac-all"} {
		if _, err := ParseRoute(raw); err != nil {
			t.Errorf("ParseRoute(%q): %v", raw, err)
		}
	}
	if _, err := ParseRoute("gateway"); err == nil {
		t.Error("queria erro na rota inventada, veio modo valido")
	}
}

func TestPacScriptGateway(t *testing.T) {
	script := PacScript(Endpoint{"socks5", "1.2.3.4", 1080}, RoutePacGateway, false)

	for _, want := range []string{`host === "discord.gg"`, `dnsDomainIs(host, ".discord.gg")`, `"SOCKS5 1.2.3.4:1080"`, `return "DIRECT"`} {
		if !strings.Contains(script, want) {
			t.Errorf("faltou %s em:\n%s", want, script)
		}
	}
	if strings.Contains(script, "; DIRECT") {
		t.Errorf("sem -fallback o PAC nao pode ter alternativa:\n%s", script)
	}
}

func TestPacScriptControle(t *testing.T) {
	script := PacScript(Endpoint{"http", "1.2.3.4", 8080}, RoutePacAll, true)

	if !strings.Contains(script, `return "PROXY 1.2.3.4:8080; DIRECT";`) {
		t.Errorf("controle deveria mandar tudo pelo proxy com alternativa:\n%s", script)
	}
	if strings.Contains(script, "dnsDomainIs") {
		t.Errorf("controle nao pode filtrar por host:\n%s", script)
	}
}

func TestPacURL(t *testing.T) {
	got := PacURL(`C:\Users\eu\AppData\Local\Cache\route.pac`)
	want := "file:///C:/Users/eu/AppData/Local/Cache/route.pac"
	if got != want {
		t.Errorf("PacURL = %q, queria %q", got, want)
	}
}
