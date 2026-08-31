# Mostra para onde o Discord conecta de verdade enquanto voce usa ele.
#
# Conta conexoes, nao destinos: SOCKS5 nao multiplexa, entao cada destino tunelado abre um TCP
# proprio para o mesmo endereco do proxy.
#
# uso: .\Validar-rota.ps1 -Canal canary -Baseline 20 -Segundos 60
#      fique parado na primeira janela, mande o arquivo na segunda.

[CmdletBinding()]
param(
    [int]$Segundos = 60,
    [int]$Baseline = 20,
    [ValidateSet("auto", "stable", "ptb", "canary")]
    [string]$Canal = "auto"
)

$binario = switch ($Canal) {
    "stable" { "Discord.exe" }
    "ptb"    { "DiscordPTB.exe" }
    "canary" { "DiscordCanary.exe" }
    default  { "Discord%.exe" }
}

$processos = Get-CimInstance Win32_Process -Filter "Name LIKE '$binario'" -ErrorAction SilentlyContinue
if (-not $processos) {
    Write-Host "Nao achei nenhum $binario aberto. Abra pelo launcher e rode de novo." -ForegroundColor Yellow
    exit 1
}

$canais = @($processos | ForEach-Object { $_.Name } | Sort-Object -Unique)
if ($canais.Count -gt 1) {
    Write-Host "Ha mais de um canal aberto ($($canais -join ', ')). Rode com -Canal ptb (ou stable)." -ForegroundColor Yellow
    exit 1
}

$rota = $null
$proxy = $null
$lista = $null
$semQuic = $false
foreach ($processo in $processos) {
    $linha = $processo.CommandLine
    if (-not $linha) { continue }

    if ($linha -match '--proxy-bypass-list=([^\s"]+)') { $lista = $matches[1] }
    if ($linha -match 'disable-quic') { $semQuic = $true }

    if ($linha -match '--proxy-server=([^\s"]+)') {
        $rota = "--proxy-server $($matches[1])"
        if ($matches[1] -match '(\d+\.\d+\.\d+\.\d+):(\d+)') { $proxy = "$($matches[1]):$($matches[2])" }
        break
    }
    if ($linha -match '--proxy-pac-url=file:///([^\s"]+)') {
        $arquivo = $matches[1] -replace '/', '\'
        $rota = "--proxy-pac-url $arquivo"
        if (Test-Path $arquivo) {
            $conteudo = Get-Content $arquivo -Raw
            if ($conteudo -match '(?:SOCKS5|PROXY|HTTPS)\s+(\d+\.\d+\.\d+\.\d+:\d+)') { $proxy = $matches[1] }
        }
        break
    }
}

if (-not $rota) {
    Write-Host "Este Discord foi aberto sem flag de proxy: nada para validar." -ForegroundColor Yellow
    Write-Host "Feche-o por inteiro e abra pelo Abrir-Discord.bat." -ForegroundColor Yellow
    exit 1
}

$pids = @($processos | ForEach-Object { $_.ProcessId })
Write-Host ""
Write-Host "  rota      $rota"
Write-Host "  proxy     $(if ($proxy) { $proxy } else { 'nao consegui extrair o endereco' })"
Write-Host "  canal     $($canais[0]), $($pids.Count) processos"
if ($lista) {
    Write-Host "  bypass    $lista"
    if ($lista -match 'storage\.googleapis\.com') {
        Write-Host "  braco     lista NOVA, com o bucket de upload" -ForegroundColor Green
    } else {
        Write-Host "  braco     lista ANTIGA, sem o bucket de upload" -ForegroundColor Yellow
    }
}
Write-Host ""

function Coletar($pids, $segundos) {
    $conexoes = @{}
    $fim = (Get-Date).AddSeconds($segundos)
    while ((Get-Date) -lt $fim) {
        foreach ($c in Get-NetTCPConnection -ErrorAction SilentlyContinue | Where-Object { $pids -contains $_.OwningProcess }) {
            $remoto = $c.RemoteAddress
            if ($remoto -in @('0.0.0.0', '::', '127.0.0.1', '::1')) { continue }

            $conexoes["$($c.LocalPort)|$remoto`:$($c.RemotePort)"] = $remoto
        }
        Start-Sleep -Milliseconds 250
    }
    return $conexoes
}

Write-Host "  NAO mexa no Discord. Medindo o normal por $Baseline segundos..." -ForegroundColor Cyan
$antes = Coletar $pids $Baseline

Write-Host "  Agora MANDE o arquivo. Olhando por $Segundos segundos..." -ForegroundColor Cyan
$depois = Coletar $pids $Segundos

if ($depois.Count -eq 0) {
    Write-Host "  Nenhuma conexao vista. O Discord estava mesmo aberto?" -ForegroundColor Yellow
    exit 1
}

function Dono($endereco, $nome) {
    switch -Regex ("$endereco $nome") {
        '^2800:3f0|^2a00:1450|^142\.250\.|^172\.217\.|^74\.125\.|^34\.|^35\.|1e100\.net|googleusercontent' { return "Google" }
        '^2606:4700|^162\.159\.|^104\.1[6-9]\.|^104\.2[0-9]\.|cloudflare' { return "Cloudflare" }
    }
    return ""
}

$destinos = @{}
foreach ($par in @(@{ Mapa = $antes; Campo = "Antes" }, @{ Mapa = $depois; Campo = "Depois" })) {
    foreach ($chave in $par.Mapa.Keys) {
        $endereco = ($chave -split '\|', 2)[1]
        if (-not $destinos.ContainsKey($endereco)) {
            $destinos[$endereco] = [pscustomobject]@{
                Endereco = $endereco
                Destino  = $par.Mapa[$chave]
                Antes    = 0
                Depois   = 0
            }
        }
        $destinos[$endereco].($par.Campo) += 1
    }
}

$linhas = foreach ($item in $destinos.Values) {
    $nome = try { (Resolve-DnsName $item.Destino -Type PTR -ErrorAction Stop | Select-Object -First 1).NameHost } catch { "" }
    [pscustomobject]@{
        Endereco = $item.Endereco
        Dono     = Dono $item.Destino $nome
        Rota     = if ($proxy -and $item.Endereco -eq $proxy) { "pelo proxy" } else { "direto" }
        Antes    = $item.Antes
        Durante  = $item.Depois
    }
}

Write-Host ""
Write-Host "  Conexoes distintas em cada janela ($Baseline s parado, $Segundos s mandando):"
$linhas | Sort-Object Rota, Endereco | Format-Table -AutoSize

$taxaAntes = { param($n) if ($Baseline -gt 0) { [math]::Round($n / $Baseline, 2) } else { 0 } }
$taxaDepois = { param($n) if ($Segundos -gt 0) { [math]::Round($n / $Segundos, 2) } else { 0 } }

$googleDireto = @($linhas | Where-Object { $_.Dono -eq "Google" -and $_.Rota -eq "direto" })
$googleNovo = @($googleDireto | Where-Object { (& $taxaDepois $_.Durante) -gt (& $taxaAntes $_.Antes) })
$peloProxy = @($linhas | Where-Object { $_.Rota -eq "pelo proxy" })
$proxySubiu = @($peloProxy | Where-Object { (& $taxaDepois $_.Durante) -gt (& $taxaAntes $_.Antes) * 1.5 })

Write-Host ""
if ($googleNovo.Count -gt 0) {
    Write-Host "  CONFIRMADO: o Google recebeu mais conexoes por fora do proxy enquanto o" -ForegroundColor Green
    Write-Host "  arquivo subia. E o bucket do anexo, e o curinga do bypass casou." -ForegroundColor Green
} elseif ($proxySubiu.Count -gt 0) {
    Write-Host "  NAO CONFIRMADO: quem recebeu conexoes novas foi o proxy, nao o Google." -ForegroundColor Red
    Write-Host "  O anexo continua subindo pelo tunel, e o curinga nao esta casando." -ForegroundColor Red
} else {
    Write-Host "  INCONCLUSIVO: nada mudou entre as duas janelas." -ForegroundColor Yellow
    if ($semQuic) {
        Write-Host "  O QUIC ja esta desligado, entao nao ha caminho de fuga por UDP: ou o" -ForegroundColor Yellow
        Write-Host "  arquivo nao subiu dentro da janela, ou reusou conexao ja aberta." -ForegroundColor Yellow
        Write-Host "  Use o teste de cronometro do protocolo, que nao depende de ver conexao." -ForegroundColor Yellow
    } else {
        Write-Host "  Isto aqui so ve TCP, e o Chromium fala QUIC (UDP) com o Google. Reabra" -ForegroundColor Yellow
        Write-Host "  assim para tirar a cegueira e rode de novo:" -ForegroundColor Yellow
        Write-Host "      .\discordgolivebypass.exe -channel canary -force -- --disable-quic" -ForegroundColor Yellow
        Write-Host "  Se ainda assim nada mudar, use o teste de cronometro do protocolo." -ForegroundColor Yellow
    }
}
