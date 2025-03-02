package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"net/http"
	"time"
)

type Cotacao struct {
	Bid string `json:"bid"`
}

func main() {
	// inicializar o banco de dados
	db, err := sql.Open("sqlite3", "./cotacoes.db")
	if err != nil {
		fmt.Printf("Erro ao abrir o banco de dados: %v\n", err)
		return
	}
	defer db.Close()

	//criar a tabela
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS cotacoes (id INTEGER PRIMARY KEY, bid TEXT, timestamp DATETIME DEFAULT CURRENT_TIMESTAMP)`)
	if err != nil {
		fmt.Printf("Erro ao criar tabela: %v\n", err)
		return
	}

	// func  que responde a rota "/"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Servidor Go no AR!")
	})

	// cria o endpoint 'cotação' que ainda não tras a resposta, por isso um 'mock' de como seria a resposta
	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		apiToConsume := "https://economia.awesomeapi.com.br/json/last/USD-BRL"

		// Fazer a requisição com um timeout configurado no contexto
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		// ja usando o ctx, configura como será a achamada pra api de cotacao
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiToConsume, nil)
		if err != nil {
			http.Error(w, "Erro ao acessar a API de cotação", http.StatusInternalServerError)
			fmt.Printf("Error ao fazer a requisição da cotação: %v/n", err)
			return
		}

		client := http.Client{}     // instancia client
		resp, err := client.Do(req) //faz o get na api pra de fato consumir o endpoint
		if err != nil {
			http.Error(w, "Erro ao requisitar API de cotação", http.StatusInternalServerError)
			fmt.Printf("Erro ao fazer requisição para API: %v\n", err)
			return
		}
		defer resp.Body.Close() // tem que ter o defer close, e pode ser aqui antes de ler o body pra não esquecer depois

		body, err := io.ReadAll(resp.Body)

		// onde será armazenado a resp em json
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			http.Error(w, "Erro ao decodificar JSON da API", http.StatusInternalServerError)
			fmt.Printf("Erro ao fazer unmarshal do JSON: %v\n", err)
			return
		}

		bid := result["USDBRL"].(map[string]interface{})["bid"].(string)

		// contexto de 10ms para inserir no db
		dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer dbCancel()

		// insere o dado da cotacao no banco
		_, err = db.ExecContext(dbCtx, "INSERT INTO cotacoes (bid) VALUES (?)", bid)
		if err != nil {
			http.Error(w, "Erro ao salvar cotação no banco de dados", http.StatusInternalServerError)
			fmt.Printf("Erro ao inserir no banco de dados: %v\n", err)
			return
		}

		// Estrutura a resposta JSON e a envia
		jsonResponse, err := json.Marshal(Cotacao{Bid: bid})
		if err != nil {
			http.Error(w, "Erro ao criar JSON de resposta", http.StatusInternalServerError)
			fmt.Printf("Erro ao formatar JSON: %v\n", err)
			return
		}

		// define o header pra tipo Json além de enviar a resposta em JSON para o cliente
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(jsonResponse)
		if err != nil {
			fmt.Printf("Erro na resposta do JSON %v\n", err)
		}
	})

	fmt.Println("Servidor na porta 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Error : %v\n", err)
	}
}
