# Client-Server API: Cotação do Dólar

Desafio Go Expert (Full Cycle): dois programas em Go que trocam informações respeitando timeouts com o pacote `context`.

- **server** (`server/server.go`): servidor HTTP na porta `8080` com o endpoint `/cotacao`. Consulta a [AwesomeAPI](https://economia.awesomeapi.com.br/json/last/USD-BRL) (timeout de **200ms**), grava a cotação no SQLite (timeout de **10ms**) e devolve um JSON `{"bid": "..."}`.
- **client** (`client/client.go`): consulta `http://localhost:8080/cotacao` (timeout de **300ms**) e salva o valor em `cotacao.txt` no formato `Dólar: {valor}`.

Quando algum timeout estoura, o erro aparece no console do respectivo programa.

## Pré-requisitos

- Go 1.26 ou superior
- Acesso à internet (para a API de câmbio)

O driver SQLite usado é o `modernc.org/sqlite`, escrito em Go puro, então não precisa de CGO nem de gcc.

## Como executar

Na raiz do projeto, baixe as dependências:

```bash
go mod download
```

### 1. Subir o servidor

```bash
cd server
go run server.go
```

O banco `cotacoes.db` é criado na pasta `server/` na primeira execução, junto com a tabela `cotacoes`.

### 2. Executar o cliente (em outro terminal)

```bash
cd client
go run client.go
```

Saída esperada:

```
Cotação salva em cotacao.txt -> Dólar: 5.5694
```

O arquivo `client/cotacao.txt` fica com o conteúdo:

```
Dólar: 5.5694
```

### Testar o endpoint direto

```bash
curl http://localhost:8080/cotacao
# {"bid":"5.5694"}
```

### Consultar as cotações gravadas

Se tiver o `sqlite3` instalado:

```bash
sqlite3 server/cotacoes.db "select * from cotacoes;"
```
