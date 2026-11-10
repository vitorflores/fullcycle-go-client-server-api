package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	_ "modernc.org/sqlite"
)

// Estrutura da resposta da API de câmbio (AwesomeAPI)
type Cotacao struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

// Resposta enviada ao cliente
type CotacaoResposta struct {
	Bid string `json:"bid"`
}

var db *sql.DB

func main() {
	var err error
	// Modo WAL deixa a escrita rápida o suficiente para caber no timeout de 10ms
	db, err = sql.Open("sqlite", "cotacoes.db?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = criarTabela(db)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/cotacao", CotacaoHandler)
	log.Println("Servidor rodando na porta 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func CotacaoHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Request iniciada")
	defer log.Println("Request finalizada")

	cotacao, err := BuscaCotacao(r.Context())
	if err != nil {
		log.Println("Erro ao buscar cotação:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = salvarCotacao(r.Context(), db, cotacao)
	if err != nil {
		log.Println("Erro ao salvar cotação no banco:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(CotacaoResposta{Bid: cotacao.USDBRL.Bid})
}

// Consome a API de câmbio com timeout máximo de 200ms
func BuscaCotacao(ctx context.Context) (*Cotacao, error) {
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Timeout de 200ms excedido ao chamar a API de câmbio")
		}
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var c Cotacao
	err = json.Unmarshal(body, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func criarTabela(db *sql.DB) error {
	_, err := db.Exec(`create table if not exists cotacoes (
		id integer primary key autoincrement,
		bid text not null,
		criado_em datetime default current_timestamp
	)`)
	return err
}

// Persiste a cotação no SQLite com timeout máximo de 10ms
func salvarCotacao(ctx context.Context, db *sql.DB, cotacao *Cotacao) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()
	stmt, err := db.PrepareContext(ctx, "insert into cotacoes(bid) values(?)")
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Timeout de 10ms excedido ao persistir no banco de dados")
		}
		return err
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, cotacao.USDBRL.Bid)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Timeout de 10ms excedido ao persistir no banco de dados")
		}
		return err
	}
	return nil
}
