# Product Views (Go + Postgres + Queue)

Um serviço minimalista e altamente performático para registrar visualizações de produtos com **baixa latência**, **alto throughput** e arquitetura **assíncrona** utilizando Go.

---

## Sumário

- [Descrição](#descrição)
- [Arquitetura](#arquitetura)
- [Tecnologias utilizadas](#tecnologias-utilizadas)
- [Fluxo de funcionamento](#fluxo-de-funcionamento)
- [Endpoints](#endpoints)
- [Como rodar](#como-rodar)
- [Variáveis de ambiente](#variáveis-de-ambiente)
- [Teste de carga](#teste-de-carga)
- [Licença](#licença)

---

## Descrição

Este serviço recebe eventos de visualização de produtos através de um endpoint HTTP leve. A requisição **não** interage diretamente com o banco de dados — ela apenas coloca o evento em uma **fila interna (channel)** e responde imediatamente com **202 Accepted**.

O processamento real é feito de forma assíncrona por **workers**, que acumulam eventos em lotes e os gravam no banco em segundo plano, reduzindo latência, contenção e pressão no Postgres.

---

## Arquitetura

```
HTTP Request
     ↓
Endpoint /events/product/view
     ↓
Fila em memória (chan)
     ↓
Workers paralelos (acumulam lotes de até 500 eventos ou 10ms)
     ↓
Batch INSERT no Postgres
```

**Principais características:**

- Handler extremamente rápido — responde sem tocar no banco.
- A fila desacopla o tráfego HTTP do banco.
- Workers fazem batch inserts para maximizar throughput no Postgres.
- Pool de conexões limitado para evitar sobrecarga no banco.
- Graceful shutdown com context cancellation.

---

## Tecnologias utilizadas

- **Go 1.22+**
- **Chi Router**
- **PostgreSQL**
- **Fila em memória via Go channels**
- **Workers assíncronos com batch insert**
- **godotenv**
- **Docker Compose**

---

## Fluxo de funcionamento

### 1. Cliente envia request:

```json
POST /events/product/view
{
  "product_id": 123
}
```

### 2. A API:

- Valida o body
- Monta um `ProductView` com timestamp UTC
- Coloca na fila (channel)
- Responde **202 Accepted** imediatamente
- Retorna **503 Service Unavailable** se a fila estiver cheia

### 3. Workers:

- Cada worker acumula eventos da fila em um batch
- Flush ocorre quando o batch atinge **500 eventos** ou a cada **10ms**
- Um único `INSERT` com múltiplas linhas é enviado ao Postgres
- Erros são logados sem travar a fila

### 4. Graceful shutdown:

- Sinais SIGINT/SIGTERM finalizam o servidor HTTP
- Workers fazem flush do batch pendente antes de encerrar

---

## Endpoints

### POST /events/product/view

Registra visualização de um produto.

**Body:**

```json
{
  "product_id": 123
}
```

**Responses:**

| Status | Descrição |
|--------|-----------|
| `202 Accepted` | Evento enfileirado com sucesso |
| `400 Bad Request` | Body inválido ou `product_id` ausente |
| `503 Service Unavailable` | Fila interna cheia |

---

### GET /health

Checagem de saúde do serviço.

**Response:** `200 OK`

---

## Como rodar

### 1. Suba o banco de dados

```bash
docker compose up -d
```

### 2. Configure o `.env` na raiz do projeto

```dotenv
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=productviews
DB_SSLMODE=disable
QUEUE_BUFFER=100000
WORKERS=20
```

### 3. Instale dependências e rode

```bash
go mod tidy
go run ./cmd/api
```

---

## Variáveis de ambiente

| Variável       | Descrição                                          |
|----------------|----------------------------------------------------|
| `DB_HOST`      | Host do Postgres                                   |
| `DB_USER`      | Usuário do Postgres                                |
| `DB_PASSWORD`  | Senha do Postgres                                  |
| `DB_NAME`      | Nome do banco                                      |
| `DB_SSLMODE`   | Modo SSL (ex: `disable`)                           |
| `QUEUE_BUFFER` | Tamanho do buffer da fila (ex: `100000`)           |
| `WORKERS`      | Número de workers assíncronos (ex: `20`)           |

---

## Teste de carga

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

O teste envia **1.000.000 requisições** com 50 usuários virtuais simultâneos para `POST /events/product/view` e reporta métricas de latência, throughput e erros ao final.

Para ajustar a carga, edite `scripts/load_test.js`:
- `vus` — usuários virtuais simultâneos (concorrência)
- `iterations` — total de requisições

---

## Banco de dados

### Acessar via linha de comando

```bash
docker compose exec postgres psql -U postgres -d productviews
```

### Consultas úteis

```sql
-- últimos registros
SELECT * FROM product_views ORDER BY id DESC LIMIT 20;

-- contar total de registros
SELECT COUNT(*) FROM product_views;

-- apagar todos os registros
TRUNCATE TABLE product_views;
```

Para sair do psql: `\q`

---

## Licença

MIT License — utilize e modifique livremente.
