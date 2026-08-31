package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type RouteMode string

const (
	RouteAll        RouteMode = "all"
	RoutePacAll     RouteMode = "pac-all"
	RoutePacGateway RouteMode = "pac-gateway"
)

const gatewayDomain = ".discord.gg"

func ParseRoute(raw string) (RouteMode, error) {
	switch mode := RouteMode(strings.ToLower(strings.TrimSpace(raw))); mode {
	case RouteAll, RoutePacAll, RoutePacGateway:
		return mode, nil
	default:
		return "", fmt.Errorf("rota invalida: %q. Use all, pac-all ou pac-gateway", raw)
	}
}

func pacProxy(e Endpoint) string {
	keyword := "PROXY"
	switch e.Scheme {
	case "socks5":
		keyword = "SOCKS5"
	case "https":
		keyword = "HTTPS"
	}
	return fmt.Sprintf("%s %s", keyword, e.address())
}

func PacScript(e Endpoint, mode RouteMode, fallbackDirect bool) string {
	proxy := pacProxy(e)
	if fallbackDirect {
		proxy += "; DIRECT"
	}

	if mode == RoutePacAll {
		return "function FindProxyForURL(url, host) {\n" +
			fmt.Sprintf("    return %q;\n", proxy) +
			"}\n"
	}

	return "function FindProxyForURL(url, host) {\n" +
		fmt.Sprintf("    if (host === %q || dnsDomainIs(host, %q)) return %q;\n", strings.TrimPrefix(gatewayDomain, "."), gatewayDomain, proxy) +
		"    return \"DIRECT\";\n" +
		"}\n"
}

func WritePac(script string) (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "DiscordGoLiveBypass", "route.pac")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(script), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func PacURL(path string) string {
	return "file:///" + strings.TrimPrefix(strings.ReplaceAll(filepath.ToSlash(path), " ", "%20"), "/")
}

func RouteArgs(e Endpoint, mode RouteMode, bypass string, fallbackDirect bool) ([]string, string, error) {
	if mode == RouteAll {
		return ProxyArgs(e, bypass, fallbackDirect), "", nil
	}

	path, err := WritePac(PacScript(e, mode, fallbackDirect))
	if err != nil {
		return nil, "", fmt.Errorf("nao consegui escrever o arquivo de rota: %w", err)
	}
	return []string{fmt.Sprintf("--proxy-pac-url=%s", PacURL(path))}, path, nil
}
