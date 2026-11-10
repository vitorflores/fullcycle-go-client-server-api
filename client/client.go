package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type CotacaoResposta struct {
	Bid string `json:"bid"`
}

func main() {
	// Timeout máximo de 300ms para receber o resultado do servidor
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Fatal("Timeout de 300ms excedido ao aguardar resposta do servidor: ", err)
		}
		log.Fatal("Erro ao requisitar cotação: ", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Fatalf("Servidor retornou status %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Fatal("Timeout de 300ms excedido ao ler resposta do servidor: ", err)
		}
		log.Fatal(err)
	}
	var cotacao CotacaoResposta
	err = json.Unmarshal(body, &cotacao)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create("cotacao.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err = f.WriteString(fmt.Sprintf("Dólar: %s", cotacao.Bid))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Cotação salva em cotacao.txt -> Dólar: %s\n", cotacao.Bid)
}
