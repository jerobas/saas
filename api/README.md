# Sweeters Backend API

Backend DDD para sistema de licenciamento com pagamento PIX via AbacatePay.

> 📚 **Documentação Completa**: Veja [DOCS_INDEX.md](DOCS_INDEX.md) para índice de toda documentação

## 🎯 Fluxo de Funcionamento

```
1. Usuário abre o aplicativo instalado
2. Tenta fazer login
3. Sistema verifica se não tem cadastro
4. Envia `POST /api/clients/register` com email e companyName
5. Backend cria Client (status: pending) e License (status: pending)
6. App recebe clientId
7. App abre tela de pagamento
8. Envia `POST /api/payments` com clientId
9. Backend integra com AbacatePay e retorna PIX (código + QR Code)
10. Usuário escaneia QR Code e paga com seu banco
11. AbacatePay envia webhook confirmando pagamento
12. Backend processa webhook e marca pagamento como pago
13. App verifica `GET /api/payments/{paymentId}/status`
14. Quando confirmado, app chama `POST /api/licenses/activate` com paymentId
15. Backend ativa a licença e retorna secretSalt
16. App salva secretSalt e usuário consegue usar o programa
17. No próximo login, app valida `POST /api/licenses/validate` com clientId
18. Backend retorna status (active/expired) + secretSalt para descriptografar dados offline
```

## 🚀 Como Rodar

### 1. Instalação de Dependências
```bash
npm install
```

### 2. Configurar Variáveis de Ambiente
```bash
cp .env.example .env
```

Edite `.env` com suas configurações:

**Opção A: Supabase (recomendado)** ⭐
```env
DATABASE_URL=postgresql://postgres-username:password@host.supabase.co:5432/postgres
```
👉 Veja [SUPABASE_SETUP.md](SUPABASE_SETUP.md) para instruções completas

**Opção B: PostgreSQL Local**
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=sweeters_db
```
Neste caso, use: `docker-compose up -d`

Também configure:
- **AbacatePay**: Configure API_KEY e WEBHOOK_SECRET

### 3. Criar Banco de Dados
```bash
npm run migrate
```

### 4. Rodar em Desenvolvimento
```bash
npm run dev
```

O servidor vai rodar em `http://localhost:3000`

### 5. Acessar Documentação
```
http://localhost:3000/api-docs
```

## 📚 Endpoints Principais

- `POST /api/clients/register` - Registrar novo cliente
- `POST /api/payments` - Criar pagamento PIX
- `GET /api/payments/{paymentId}/status` - Check pagamento
- `POST /api/webhooks/payment` - Webhook AbacatePay (automático)
- `POST /api/licenses/activate` - Ativar após pagamento
- `POST /api/licenses/validate` - Validar se está ativa

## 🏗️ Arquitetura em Camadas

```
src/
├── domain/                    # Lógica de negócio
│   ├── licenses/
│   │   ├── entities/         # Client, License
│   │   └── repositories/     # Interfaces de dados
│   └── payments/
│       ├── entities/         # Payment
│       └── repositories/
├── usecases/                  # Casos de uso
│   ├── register-client-usecase.js
│   ├── create-payment-usecase.js
│   ├── validate-license-usecase.js
│   └── active-license-usecase.js
├── infra/
│   ├── database/             # Configuração TypeORM
│   ├── http/
│   │   ├── routes/           # Rotas Express
│   │   ├── controllers/      # Controllers
│   │   ├── middlewares/      # Middleware de erro
│   │   ├── errors/           # AppError
│   │   └── swagger/          # Documentação
│   └── services/             # AbacatePayService
└── main.js                    # Entry point
```

## 🔐 Segurança

- **Validação de Webhook**: Todos os webhooks são validados com HMAC-SHA256
- **SecretSalt**: Gerado por cliente para criptografia local offline
- **Status de Licença**: Três estados (pending, active, expired)
- **Offline Grace Period**: Modo offline com validade configurável (padrão: 72h)

## 🌐 Integração AbacatePay

Este backend se integra com [AbacatePay](https://abacatepay.com.br) para processar pagamentos PIX.

### Fluxo de Integração:
1. **Criar Pagamento**: Backend chama AbacatePay para gerar PIX
2. **Webhook**: AbacatePay notifica backend quando usuário paga
3. **Ativar Licença**: Backend ativa licença após confirmação

### Variáveis Necessárias:
```env
ABACATEPAY_API_KEY=sua_chave_api
ABACATEPAY_WEBHOOK_SECRET=seu_segredo_webhook
ABACATEPAY_BASE_URL=https://api.abacatepay.com.br
```

## 📝 Exemplos de Requisição

### 1. Registrar Cliente
```bash
curl -X POST http://localhost:3000/api/clients/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "usuario@example.com",
    "companyName": "Minha Empresa"
  }'
```

### 2. Criar Pagamento
```bash
curl -X POST http://localhost:3000/api/payments \
  -H "Content-Type: application/json" \
  -d '{
    "clientId": "uuid-do-cliente",
    "amount": 99.90
  }'
```

### 3. Validar Licença (No login da app)
```bash
curl -X POST http://localhost:3000/api/licenses/validate \
  -H "Content-Type: application/json" \
  -d '{
    "clientId": "uuid-do-cliente"
  }'
```

## 🛠️ Troubleshooting

**Erro de conexão Database:**
- Verifique credenciais em `.env`
- Certifique-se que PostgreSQL está rodando

**Erro AbacatePay:**
- Verifique API_KEY nas variáveis ambiente
- Confira se webhook URL está corretamente configurada

**Licença não ativa:**
- Verifique se pagamento foi confirmado
- Confira se `activate` foi chamado após payment.status = 'paid'

## 📚 Documentação Completa

Veja [DOCS_INDEX.md](DOCS_INDEX.md) para:
- ✅ [SETUP_CHECKLIST.md](SETUP_CHECKLIST.md) - Checklist passo a passo
- ✅ [SUPABASE_SETUP.md](SUPABASE_SETUP.md) - Setup Supabase detalhado
- ✅ [ARCHITECTURE.md](ARCHITECTURE.md) - Arquitetura e diagramas
- ✅ [CLIENT_INTEGRATION.md](CLIENT_INTEGRATION.md) - Como integrar seu app
- ✅ [PRODUCTION.md](PRODUCTION.md) - Deploy e produção
- ✅ [QUICKSTART.md](QUICKSTART.md) - 7 passos rápidos

## 📄 Licença

MIT
