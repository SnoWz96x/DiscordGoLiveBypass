package main

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"syscall"
)

// WSAEACCES. O Windows devolve esse código quando quem recusou o socket foi a própria
// máquina, não a rede: antivírus, firewall ou uma camada de filtragem instalada no meio.
// O número só existe no Windows, então em outros sistemas a comparação simplesmente nunca
// bate e o EACCES logo abaixo cobre o caso equivalente.
const winSocketForbidden = syscall.Errno(10013)

// A mensagem crua do net/http carrega a URL inteira e o IP resolvido, que enchem a caixa de
// erro sem dizer o que fazer. Aqui sobra a causa e, quando dá, o próximo passo.
func DescribeNetwork(err error) string {
	if err == nil {
		return ""
	}

	if errors.Is(err, winSocketForbidden) || errors.Is(err, syscall.EACCES) {
		return "a conexao foi bloqueada pela sua propria maquina, antes de sair para a internet. Quase sempre e antivirus ou firewall barrando o DiscordGoLiveBypass.exe, que nao tem assinatura digital: libere o programa na lista de excecoes e rode de novo"
	}

	var dns *net.DNSError
	if errors.As(err, &dns) {
		return fmt.Sprintf("nao consegui resolver o endereco %s, o que costuma ser DNS ou internet fora do ar", dns.Name)
	}

	var timeout net.Error
	if errors.As(err, &timeout) && timeout.Timeout() {
		return "a resposta demorou demais, a conexao pode estar instavel"
	}

	if errors.Is(err, syscall.ECONNREFUSED) {
		return "a conexao foi recusada do outro lado"
	}

	var wrapped *url.Error
	if errors.As(err, &wrapped) {
		return wrapped.Err.Error()
	}
	return err.Error()
}
