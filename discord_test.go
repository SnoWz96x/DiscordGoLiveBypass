package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewerThan(t *testing.T) {
	cases := []struct {
		name string
		a, b string
		want bool
	}{
		{"numero maior no ultimo campo", "app-1.0.9184", "app-1.0.998", true},
		{"numero menor no ultimo campo", "app-1.0.998", "app-1.0.9184", false},
		{"versoes iguais", "app-1.0.9253", "app-1.0.9253", false},
		{"campo a mais desempata", "app-1.0.9253.1", "app-1.0.9253", true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := newerThan(versionOf(c.a), versionOf(c.b)); got != c.want {
				t.Errorf("newerThan(%s, %s) = %v, queria %v", c.a, c.b, got, c.want)
			}
		})
	}
}

func TestVersionOf(t *testing.T) {
	for _, name := range []string{"app-1.0.beta", "app-", "Update.exe"} {
		if got := versionOf(name); got != nil {
			t.Errorf("versionOf(%q) = %v, queria nil", name, got)
		}
	}
}

func TestChannelForBinary(t *testing.T) {
	if got := ChannelForBinary(`C:\Discord\app-1.0.1\DiscordPTB.exe`); got.Key != "ptb" {
		t.Errorf("canal do DiscordPTB.exe = %q, queria ptb", got.Key)
	}
	if got := ChannelForBinary(`C:\Outro\Vesktop.exe`); got.Key != "custom" {
		t.Errorf("canal do Vesktop.exe = %q, queria custom", got.Key)
	}
}

func TestLocateDiscord(t *testing.T) {
	channel := channels["stable"]

	t.Run("prefere a completa mesmo sendo mais velha", func(t *testing.T) {
		root := t.TempDir()
		t.Setenv("LOCALAPPDATA", root)

		completa := filepath.Join(root, channel.Folder, "app-1.0.998")
		install(t, completa, channel.Binary, true)
		install(t, filepath.Join(root, channel.Folder, "app-1.0.9184"), channel.Binary, false)

		binary, err := LocateDiscord(channel)
		if err != nil {
			t.Fatalf("LocateDiscord: %v", err)
		}
		if want := filepath.Join(completa, channel.Binary); binary != want {
			t.Errorf("escolheu %s, queria %s", binary, want)
		}
	})

	t.Run("recusa quando nenhuma esta completa", func(t *testing.T) {
		root := t.TempDir()
		t.Setenv("LOCALAPPDATA", root)

		install(t, filepath.Join(root, channel.Folder, "app-1.0.9184"), channel.Binary, false)

		if binary, err := LocateDiscord(channel); err == nil {
			t.Errorf("queria erro, veio %s", binary)
		}
	})
}

// Instalação incompleta é a que tem o executável e não tem os arquivos do Chromium.
func install(t *testing.T, dir, binary string, complete bool) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	names := []string{binary}
	if complete {
		names = append(names, runtimeFiles...)
	}
	for _, name := range names {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
