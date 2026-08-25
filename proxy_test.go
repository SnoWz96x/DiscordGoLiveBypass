package main

import "testing"

func TestParseProxy(t *testing.T) {
	cases := []struct {
		raw  string
		want Endpoint
		ok   bool
	}{
		{"socks5://1.2.3.4:1080", Endpoint{"socks5", "1.2.3.4", 1080}, true},
		{"  HTTP://Proxy.Exemplo.Com:8080  ", Endpoint{"http", "proxy.exemplo.com", 8080}, true},
		{"1.2.3.4:1080", Endpoint{}, false},
		{"socks4://1.2.3.4:1080", Endpoint{}, false},
		{"socks5://1.2.3.4", Endpoint{}, false},
		{"socks5://1.2.3.4:70000", Endpoint{}, false},
		{"socks5://1.2.3.4:1080/caminho", Endpoint{}, false},
	}

	for _, c := range cases {
		t.Run(c.raw, func(t *testing.T) {
			got, ok := ParseProxy(c.raw)
			if ok != c.ok {
				t.Fatalf("ParseProxy(%q) aceitou = %v, queria %v", c.raw, ok, c.ok)
			}
			if ok && got != c.want {
				t.Errorf("ParseProxy(%q) = %v, queria %v", c.raw, got, c.want)
			}
		})
	}
}
