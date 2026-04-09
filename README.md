# 🟦 Product Views (Go + Postgres + Queue)

Um serviço minimalista, altamente performático e focado em registrar
visualizações de produtos com **latência baixíssima**, **alto
throughput** e arquitetura **assíncrona** utilizando Go.

Este projeto nasceu como um desafio:\
**até onde é possível escalar um serviço simples em uma instância
extremamente limitada (t3.micro, 1GB RAM, 2 vCPUs)?**

O resultado alcançado em testes de carga:\
**1.000.000 requisições processadas em 63 segundos**\
**Zero erros**\
**\~15.695 requests/s (\~940k/min)**\
Latência média: **\~15ms**

Tudo isso em uma máquina minúscula --- graças a uma arquitetura
eficiente.

---

## 📌 Sumário

- [Descrição](#descrição)
- [Arquitetura](#arquitetura)
- [Tecnologias utilizadas](#tecnologias-utilizadas)
- [Fluxo de funcionamento](#fluxo-de-funcionamento)
- [Resultados de performance](#resultados-de-performance)
- [Endpoints](#endpoints)
- [Como rodar](#como-rodar)
- [Variáveis de ambiente](#variáveis-de-ambiente)
- [Estrutura do projeto](#estrutura-do-projeto)
- [Próximos passos](#próximos-passos)
- [Licença](#licença)

---

## 📘 Descrição

Este serviço recebe eventos de visualização de produtos através de um
endpoint HTTP extremamente leve.\
A requisição **não** interage diretamente com o banco de dados --- ela
apenas coloca o evento em uma **fila interna (channel)** e responde
imediatamente com **202 Accepted**.

O processamento real é feito de forma assíncrona por **workers**, que
consomem a fila e gravam no banco em segundo plano.

Essa abordagem reduz drasticamente: - latência, - bloqueios, -
contenção, - pressão no Postgres, - e melhora o throughput global.

---

## ⚙ Arquitetura

A arquitetura foi desenhada com foco em **simplicidade e eficiência**:

    HTTP Request
         ↓
     Endpoint /events/product/view
         ↓
       Fila em memória (chan)
         ↓
    Workers paralelos
         ↓
    Persistência no Postgres

### **Principais características**

- Handler extremamente rápido (**5--30µs**).
- A fila desacopla o tráfego HTTP do banco.
- Workers controlam o ritmo de gravação.
- O banco trabalha de forma estável e previsível.
- Suporte a graceful shutdown.
- Baixo consumo de CPU e memória.
- Ideal para serviços de alta escala e eventos.

---

## 🛠 Tecnologias utilizadas

- **Go 1.22+**
- **Chi Router**
- **PostgreSQL**
- **Fila em memória via Go channels**
- **Workers assíncronos**
- **godotenv**
- **Context API nativa do Go**

---

## 🔄 Fluxo de funcionamento

### 1. Cliente envia request:

```json
POST /events/product/view
{
  "product_id": 123
}
```

### 2. A API:

- Valida\
- Monta um evento\
- Coloca na fila\
- Responde **202 Accepted** imediatamente

### 3. Workers:

- Ouvem a fila\
- Cada evento é gravado no Postgres\
- Erros são logados, sem travar a fila

### 4. Graceful shutdown:

- Sinais SIGINT/SIGTERM finalizam o servidor\
- Workers aguardam a fila esvaziar

---

## 🚀 Resultados de performance

Rodado em uma AWS **t3.micro (1GB RAM / 2 vCPUs)**:

### **Rodada final de benchmark**

- **1.000.000 requisições**
- **63 segundos**
- **0 erros**
- **15.695 req/s**
- **941.700 req/min**
- **latência média: 15ms**
- Nenhum evento perdido

---

## 📡 Endpoints

### **POST /events/product/view**

Registra visualização de um produto.

**Body:**

```json
{
  "product_id": 123
}
```

**Response:**

    202 Accepted

---

### **GET /health**

Checagem simples de saúde.

**Response:**

    200 OK

---

## ▶ Como rodar

### 1. Clone o repositório

```bash
git clone https://github.com/seu-user/product-views
cd product-views
```

### 2. Crie o arquivo `.env`

```dotenv
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=productviews
QUEUE_BUFFER_SIZE=100000
WORKER_COUNT=10
HTTP_PORT=3000
```

### 3. Instale dependências

```bash
go mod tidy
```

### 4. Rode

```bash
go run ./cmd/api
```

---

## 📁 Estrutura do projeto

    /cmd/api
        main.go

    /internal
        /httpapi
            handlers.go
            router.go

        /queue
            queue.go

        /repository
            db.go
            product_views_repo.go

        /domain
            product.go
            product_view.go

        /config
            config.go

## 🧪 Teste de carga

### 1. Instale o k6

```bash
brew install k6
```

### 2. Suba o banco e a aplicação

```bash
docker compose up -d
go run ./cmd/api
```

### 3. Rode o teste

```bash
k6 run scripts/load_test.js
```

O teste envia **1.000.000 requisições** com 100 usuários virtuais simultâneos para `POST /events/product/view` e reporta métricas de latência, throughput e erros ao final.

Para ajustar a carga, edite `scripts/load_test.js`:
- `vus` — usuários virtuais simultâneos (concorrência)
- `iterations` — total de requisições

---

## 📄 Licença

MIT License --- utilize e modifique livremente.
