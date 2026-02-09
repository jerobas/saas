# 🚀 SaaS Licensing Backend

Backend DDD para sistema de licenciamento com pagamento PIX via AbacatePay.

Stack: **Node.js 22.16.0** | **Express** | **PostgreSQL (Supabase)** | **TypeORM** | **Ed25519**

## 📋 Índice

- [Visão Geral](#visão-geral)
- [Arquitetura](#arquitetura)
- [Requisitos](#requisitos)
- [Instalação](#instalação)
- [Configuração](#configuração)
- [Como Rodar](#como-rodar)
- [API Endpoints](#api-endpoints)
- [Fluxo de Pagamento](#fluxo-de-pagamento)
- [Estrutura do Projeto](#estrutura-do-projeto)
- [Testes](#testes)
- [Deploy](#deploy)

## 🎯 Visão Geral

Sistema completo de licenciamento que integra:

- ✅ **Criação de usuários** - Registro com email, nome, CPF/CNPJ, telefone
- ✅ **Integração AbacatePay** - PIX para pagamento
- ✅ **Webhook de confirmação** - Automático quando usuário paga
- ✅ **Geração de tokens** - Licença assinada com Ed25519
- ✅ **Renovação de assinatura** - Usuários existentes renovam licença
- ✅ **Verificação de status** - App consulta licença do usuário
- ✅ **Sem Email/SMS** - Experiência limpa para o usuário

## 🏗️ Arquitetura

```
┌─────────────────────────────────────────┐
│         Node.js 22.16.0 Express         │
├─────────────────────────────────────────┤
│         Domain-Driven Design (DDD)      │
├─────────────────────────────────────────┤
│  Domain Layer (Users, Payments)         │
│  ├── Entities                           │
│  ├── Repositories                       │
│  └── Use Cases                          │
├─────────────────────────────────────────┤
│  Infrastructure Layer                   │
│  ├── HTTP (Express Controllers/Routes)  │
│  ├── Database (TypeORM + Supabase)      │
│  └── Services (AbacatePay, License)     │
├─────────────────────────────────────────┤
│         PostgreSQL (Supabase)           │
└─────────────────────────────────────────┘
```

## ⚙️ Requisitos

- **Node.js 22.16.0** (com npm)
- **PostgreSQL** (Supabase Cloud)
- **OpenSSL** (para chaves criptográficas)

## 🔧 Instalação

### 1. Dependências
```bash
npm install
```

### 2. Gerar Chaves Ed25519
```bash
bash generate-keys.sh
```

Cria:
- `src/license/private.pem` - Chave privada (⚠️ adicionar ao .gitignore)
- `src/license/public.pem` - Chave pública

### 3. Configurar Ambiente
```bash
cp .env.example .env
```

## 🔐 Configuração (.env)

```env
# Server
PORT=3000
NODE_ENV=development

# Database (Supabase PostgreSQL)
DATABASE_URL=postgresql://user:password@host:port/database

# AbacatePay
ABACATEPAY_BASE_URL=https://api.abacatepay.com/v1
ABACATEPAY_API_KEY=abc_dev_...
ABACATEPAY_WEBHOOK_SECRET=webhook_secret_...

# License (em centavos)
LICENSE_PRICE=50000          # R$ 500,00
LICENSE_DURATION_DAYS=365    # 1 ano

# SMTP (Opcional - se precisar enviar emails)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=seu-email@gmail.com
SMTP_PASSWORD=app-password

# Security
JWT_SECRET=sua-chave-super-secreta
ENCRYPTION_MASTER_KEY=sua-chave-mestre
```

## ▶️ Como Rodar

### Desenvolvimento
```bash
npm run dev
```

Servidor: `http://localhost:3000`
Swagger: `http://localhost:3000/api-docs`

### Produção
```bash
npm start
```

## 📡 API Endpoints

### 1️⃣ Criar Usuário + Gerar PIX
```http
POST /api/create
Content-Type: application/json

{
  "email": "usuario@example.com",
  "name": "João Silva",
  "taxId": "12345678901",
  "cellphone": "11999999999"
}

✅ Response (201):
{
  "success": true,
  "data": {
    "userId": "550e8400-e29b-41d4-a716-446655440000",
    "paymentId": "uuid",
    "brCode": "00020101021226950014br.gov.bcb.pix...",
    "brCodeBase64": "data:image/png;base64,iVBORw0KGgo...",
    "expiresAt": "2026-02-10T12:00:00Z",
    "amount": 500
  }
}
```

### 2️⃣ Renovar Assinatura
```http
POST /api/renew
Content-Type: application/json

{
  "userId": "550e8400-e29b-41d4-a716-446655440000"
}

✅ Response (200): Mesmo formato do endpoint de criar
```

### 3️⃣ Webhook de Pagamento
```http
POST /api/webhooks/payment
Content-Type: application/json
X-AbacatePay-Signature: signature_value

{
  "event": "payment.confirmed",
  "data": {
    "id": "pix_transaction_id"
  }
}

Events:
- payment.confirmed  → Licença ativada
- payment.cancelled  → PIX cancelado
- payment.expired    → PIX expirou

✅ Response (200):
{
  "success": true,
  "data": {
    "event": "payment.confirmed",
    "paymentId": "uuid",
    "userId": "uuid",
    "status": "PAID",
    "message": "Pagamento confirmado e licença ativada"
  }
}
```

### 4️⃣ Verificar Status da Licença
```http
GET /api/users/{userId}/license

✅ Response (200):
{
  "success": true,
  "data": {
    "userId": "uuid",
    "email": "usuario@example.com",
    "licenseActive": true,
    "licenseExpiresAt": "2027-02-09T11:00:00Z",
    "licenseToken": "eyJ0eXAiOiJKV1QiLCJhbGc...",
    "daysRemaining": 365
  }
}

❌ Response (404): Usuário não encontrado
❌ Response (403): Licença não está ativa
```

### 5️⃣ Simular Pagamento (DEV ONLY)
```http
POST /api/dev/simulate
Content-Type: application/json

{
  "paymentId": "uuid"
}

⚠️ Requer: NODE_ENV=development
```

## 💳 Fluxo de Pagamento

```
┌──────────────────────┐
│ 1. POST /api/create  │  Cria User + AbacatePay Customer + PIX
└──────────────────────┘
         │
         ▼
┌──────────────────────┐
│ 2. Exibe QR Code     │  App exibe PIX para escanear
└──────────────────────┘
         │
         ▼
┌──────────────────────┐
│ 3. Usuário Paga PIX  │  Escaneia + paga via banco
└──────────────────────┘
         │
         ▼
┌──────────────────────────┐
│ 4. AbacatePay Confirma   │  Envia evento de pagamento
└──────────────────────────┘
         │
         ▼
┌──────────────────────────┐
│ 5. POST /webhooks/payment│  Backend recebe confirmação
└──────────────────────────┘
         │
         ▼
┌──────────────────────────┐
│ 6. Ativa Licença         │  Salva licenseActive=true
└──────────────────────────┘
         │
         ▼
┌──────────────────────────┐
│ 7. GET /users/{id}/license
         │  App verifica status
└──────────────────────────┘
         │
         ▼
┌──────────────────────────┐
│ 8. Recebe Token + Valida │  App armazena token localmente
└──────────────────────────┘
```

## 📁 Estrutura do Projeto

```
src/
├── main.js                              # Entry point
├── domain/                              # Lógica de negócio (DDD)
│   ├── users/
│   │   ├── entities/
│   │   │   └── user.entity.js          # Entidade User
│   │   ├── repositories/
│   │   │   └── user.repository.js      # CRUD + Queries
│   │   └── usecases/
│   │       ├── create-user-usecase.js
│   │       ├── renew-subscription-usecase.js
│   │       └── get-license-status-usecase.js
│   └── payments/
│       ├── entities/
│       │   └── payment.entity.js       # Entidade Payment
│       ├── repositories/
│       │   └── payment.repository.js   # CRUD + Queries
│       └── usecases/
│           ├── process-payment-webhook-usecase.js
│           └── simulate-payment-usecase.js
├── infra/
│   ├── database/
│   │   ├── data-source.js              # TypeORM + Supabase
│   │   └── migrations/
│   ├── http/
│   │   ├── controllers/                # Handlers HTTP
│   │   │   ├── create-user.controller.js
│   │   │   ├── renew-subscription.controller.js
│   │   │   ├── webhook-payment.controller.js
│   │   │   ├── simulate-payment.controller.js
│   │   │   └── get-license-status.controller.js
│   │   ├── routes/
│   │   │   └── index.js               # Express Router
│   │   ├── swagger/
│   │   │   └── setup.js               # Documentação
│   │   ├── middlewares/
│   │   │   └── error-handler.js
│   │   └── errors/
│   │       └── app-error.js
│   └── services/
│       ├── abacatepay.service.js      # Integração AbacatePay
│       ├── license.service.js         # Geração de tokens Ed25519
│       └── email.service.js           # SMTP (Nodemailer)
└── license/
    ├── private.pem                    # Chave privada Ed25519
    └── public.pem                     # Chave pública Ed25519
```

## 📊 Entidades

### User
```javascript
{
  id: UUID,
  email: string (unique),
  name: string,
  taxId: string,
  cellphone: string,
  abacatePayCustomerId: string,
  licenseActive: boolean,
  licenseExpiresAt: timestamp,
  licenseToken: string,
  createdAt: timestamp,
  updatedAt: timestamp
}
```

### Payment
```javascript
{
  id: UUID,
  userId: UUID (FK to User),
  abacatePayCustomerId: string,
  abacatePayPixId: string,
  amount: bigint (em centavos),
  status: PENDING | PAID | CANCELLED | EXPIRED,
  brCode: string,
  brCodeBase64: string,
  expiresAt: timestamp,
  error: string (nullable),
  createdAt: timestamp,
  updatedAt: timestamp
}
```

## 🧪 Testes

### Testar Geração de Licenças

```bash
# Teste básico
node test-license.js

# Teste completo (6 cenários)
node test-complete.js
```

**Arquivos de teste:**
- `test-license.js` - Teste simples e rápido
- `test-payloads.js` - Dados de teste para diferentes cenários
- `test-complete.js` - Suite completa com validações

## 🔒 Segurança

### Chaves Criptográficas
- **Algoritmo:** Ed25519 (NIST recomendado)
- **Geração:** OpenSSL
- **Armazenamento:** `src/license/*.pem` (⚠️ adicionar ao .gitignore)

### Validação de Webhook
- **Header:** X-AbacatePay-Signature
- **Método:** HMAC-SHA256

### Tokens de Licença
- **Formado:** Payload JSON comprimido com GZIP + assinatura Ed25519
- **Conteúdo:** userId, email, issuedAt, expiresAt
- **Verificação:** Offline com chave pública

## 🚀 Deploy

### Heroku/Railway/Render
```bash
# Build automático detecta Node.js
# Scripts rodados:
# - npm install
# - npm start

# Variáveis obrigatórias em produção:
NODE_ENV=production
DATABASE_URL=postgresql://...
ABACATEPAY_API_KEY=...
ABACATEPAY_WEBHOOK_SECRET=...
JWT_SECRET=gerar-chave-forte-256
```

### Docker
```dockerfile
FROM node:22.16.0-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
COPY src/license/*.pem ./src/license/
EXPOSE 3000
CMD ["npm", "start"]
```

## 📝 Licença

MIT
