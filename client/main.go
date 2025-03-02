package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Cotacao struct {
	Bid string `json:"bid"`
}

func main() {
	// contexto com timeout de 300ms para a requisição
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao", nil)
	if err != nil {
		fmt.Printf("Erro ao criar a requisição: %v\n", err)
		return
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Erro ao fazer a requisição: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Erro ao ler o corpo da resposta: %v\n", err)
		return
	}

	var cotacao Cotacao
	if err := json.Unmarshal(body, &cotacao); err != nil {
		fmt.Printf("Erro ao decodificar o JSON: %v\n", err)
		return
	}

	file, err := os.OpenFile("cotacao.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Erro ao abrir arquivo: %v\n", err)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("Erro ao fechar arquivo: %v\n", err)
		}
	}(file)

	if _, err := file.WriteString(fmt.Sprintf("Dólar: %s\n", cotacao.Bid)); err != nil {
		fmt.Printf("Erro ao escrever no arquivo: %v\n", err)
		return
	}

	fmt.Println("Cotação salva em cotacao.txt")
}
