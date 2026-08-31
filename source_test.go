package main

import (
	"testing"
	"time"
)

func TestParseExcluded(t *testing.T) {
	got := ParseExcluded(" br , AR, x, 123, ")
	want := map[string]bool{"BR": true, "AR": true}

	if len(got) != len(want) {
		t.Fatalf("ParseExcluded = %v, queria %v", got, want)
	}
	for code := range want {
		if !got[code] {
			t.Errorf("faltou %s em %v", code, got)
		}
	}

	if empty := ParseExcluded(""); len(empty) != 0 {
		t.Errorf("ParseExcluded(\"\") = %v, queria vazio", empty)
	}
}

// Um descartado por filtro, e os dois que sobram na ordem em que serao testados.
func TestRankFreeProxies(t *testing.T) {
	body := `{"proxies":[
		{"proxy":"socks5://1.1.1.1:1080","alive":true,"uptime":99,"timeout":300,"ip_data":{"countryCode":"DE"}},
		{"proxy":"socks5://2.2.2.2:1080","alive":true,"uptime":95,"timeout":200,"ip_data":{"countryCode":"BR"}},
		{"proxy":"socks5://3.3.3.3:4145","alive":true,"uptime":99,"timeout":100,"ip_data":{"countryCode":"US"}},
		{"proxy":"socks5://4.4.4.4:1080","alive":false,"uptime":99,"timeout":100,"ip_data":{"countryCode":"US"}},
		{"proxy":"socks5://5.5.5.5:1080","alive":true,"uptime":80,"timeout":100,"ip_data":{"countryCode":"US"}},
		{"proxy":"socks5://6.6.6.6:1080","alive":true,"uptime":99,"timeout":9000,"ip_data":{"countryCode":"US"}},
		{"proxy":"socks5://7.7.7.7:1080","alive":true,"uptime":100,"timeout":800,"ip_data":{"countryCode":"NL"}}
	]}`

	ranked, err := rankFreeProxies(body, map[string]bool{"BR": true})
	if err != nil {
		t.Fatalf("rankFreeProxies: %v", err)
	}

	want := []string{"socks5://7.7.7.7:1080", "socks5://1.1.1.1:1080"}
	if len(ranked) != len(want) {
		t.Fatalf("sobraram %d candidatas (%v), queria %v", len(ranked), ranked, want)
	}
	for i, endpoint := range ranked {
		if endpoint.String() != want[i] {
			t.Errorf("posicao %d = %s, queria %s", i, endpoint, want[i])
		}
	}
}

func TestRankFreeProxiesBadJSON(t *testing.T) {
	if _, err := rankFreeProxies("<html>nao sou json</html>", nil); err == nil {
		t.Error("queria erro, veio lista")
	}
}

func TestRankProbed(t *testing.T) {
	pool := []probeResult{
		{Endpoint{"socks5", "1.1.1.1", 1080}, 2000 * time.Millisecond},
		{Endpoint{"socks5", "2.2.2.2", 1080}, 300 * time.Millisecond},
		{Endpoint{"socks5", "3.3.3.3", 1080}, 900 * time.Millisecond},
	}
	asked := map[string]bool{"socks5://2.2.2.2:1080": true}

	got := rankProbed(pool, preferredLatency, asked)
	if len(got) != 1 || got[0].endpoint.Host != "3.3.3.3" {
		t.Fatalf("sob o limite preferido = %v, queria so a de 900 ms", got)
	}

	got = rankProbed(pool, probeTimeout, asked)
	want := []string{"3.3.3.3", "1.1.1.1"}
	if len(got) != len(want) {
		t.Fatalf("sem limite apertado sobraram %d, queria %v", len(got), want)
	}
	for i, result := range got {
		if result.endpoint.Host != want[i] {
			t.Errorf("posicao %d = %s, queria %s", i, result.endpoint.Host, want[i])
		}
	}
}
