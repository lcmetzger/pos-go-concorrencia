package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func TestAuctionFlow(t *testing.T) {
	// faz a leitura das variáveis de ambiente para descobrir qual o tempo de timeout
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error trying to load env variables")
		return
	}

	// Faz a leitura da variável de ambiente para descobrir qual o tempo de timeout
	auctionTimeout := os.Getenv("AUCTION_TIMEOUT")
	timeoutDuration, err := time.ParseDuration(auctionTimeout)
	if err != nil {
		log.Fatal("Ocorreu um erro ao converter o timeout do arquivo .env", err)
	}

	// Dados para a requisição POST /auction
	auctionData := map[string]interface{}{
		"product_name": "product name",
		"category":     "categoria",
		"description":  "descrição descrição descrição descrição descrição descrição descrição descrição ",
		"condition":    0,
	}

	// Fazer a requisição POST para /auction
	err = makePostRequest("http://localhost:8080/auction", auctionData)
	if err != nil {
		t.Fatalf("Erro ao fazer POST para /auction: %v", err)
	}

	// Obter o ID do leilão criado
	auctionGetResp, err := makeGetRequest("http://localhost:8080/auction?status=0")
	if err != nil {
		t.Fatalf("Erro ao fazer GET para /auction: %v", err)
	}

	// recupera o ID do ultimo leilão criado
	if len(auctionGetResp) == 0 {
		t.Fatalf("Nenhum leilão encontrado")
	}
	auctionID := auctionGetResp[len(auctionGetResp)-1]["id"].(string)

	// Dados para a requisição POST /bid
	bidData := map[string]interface{}{
		"user_id":    auctionID,
		"auction_id": auctionID,
		"amount":     1000,
	}

	// Fazer a requisição POST para /bid
	err = makePostRequest("http://localhost:8080/bid", bidData)
	if err != nil {
		t.Fatalf("Erro ao fazer POST para /bid: %v", err)
	}

	// Fazer a requisição GET para /bid/:auctionId e contar os registros
	initialCount, err := countBids("http://localhost:8080/bid/" + auctionID)
	if err != nil {
		t.Fatalf("Erro ao fazer GET para /bid/:auctionId: %v", err)
	}

	// Aguardar o fechamento do leilão
	time.Sleep(timeoutDuration + 5)

	// Fazer a requisição POST para /bid
	err = makePostRequest("http://localhost:8080/bid", bidData)
	if err != nil {
		t.Fatalf("Erro ao fazer POST para /bid: %v", err)
	}

	// Fazer a requisição GET para /bid/:auctionId novamente e contar os registros
	finalCount, err := countBids("http://localhost:8080/bid/" + auctionID)
	if err != nil {
		t.Fatalf("Erro ao fazer GET para /bid/:auctionId: %v", err)
	}

	// Fazer a asserção dos registros retornados
	if initialCount != finalCount {
		t.Fatalf("A contagem de registros não corresponde: inicial %d, final %d", initialCount, finalCount)
	}
}

func makePostRequest(url string, data map[string]interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("erro ao converter dados para JSON: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("erro ao fazer requisição POST: %v", err)
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("erro ao ler resposta: %v", err)
	}

	return nil
}

func countBids(url string) (int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("erro ao fazer requisição GET: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("erro ao ler resposta: %v", err)
	}

	var bids []map[string]interface{}
	err = json.Unmarshal(body, &bids)
	if err != nil {
		return 0, fmt.Errorf("erro ao converter resposta para JSON: %v", err)
	}

	return len(bids), nil
}

// Busca os registros de um endpoint
func makeGetRequest(url string) ([]map[string]interface{}, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result []map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
