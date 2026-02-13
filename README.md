# SaaS Project

## Arquitetura do Projeto

Este projeto é uma aplicação SaaS (Software as a Service) composta por um backend desenvolvido em Node.js e um frontend desenvolvido em React. A arquitetura é baseada em uma abordagem modular e utiliza RabbitMQ para gerenciamento de filas e comunicação assíncrona.

### Estrutura do Projeto

#### Backend (Pasta `api/`)
O backend é responsável por gerenciar a lógica de negócios, comunicação com o banco de dados e integração com serviços externos. A estrutura do backend é organizada da seguinte forma:

- **docker-compose.yml**: Configuração para serviços Docker, incluindo RabbitMQ e outros serviços necessários.
- **package.json**: Gerenciamento de dependências e scripts do Node.js.
- **src/**: Contém o código-fonte principal do backend.
  - **domain/**: Contém a lógica de domínio dividida em módulos, como `payments` e `users`.
    - **entities/**: Define as entidades do domínio, como `payment.entity.js` e `user.entity.js`.
    - **repositories/**: Gerencia a persistência de dados, como `payment.repository.js` e `user.repository.js`.
    - **usecases/**: Contém os casos de uso, como `process-payment-webhook-usecase.js` e `simulate-payment-usecase.js`.
  - **infra/**: Contém a infraestrutura do projeto.
    - **database/**: Configuração e conexão com o banco de dados.
    - **http/**: Gerencia a camada de API REST.
      - **controllers/**: Controladores para lidar com as requisições HTTP.
      - **errors/**: Gerenciamento de erros personalizados.
      - **middlewares/**: Middlewares para tratamento de requisições.
      - **routes/**: Configuração das rotas da API.
      - **swagger/**: Configuração da documentação da API.
    - **queue/**: Configuração e consumidores do RabbitMQ.
      - **rabbitmq.js**: Configuração do cliente RabbitMQ.
      - **user-strategy-consumer.js**: Consumidor para estratégias de criação de usuários e pagamentos Pix.
    - **services/**: Serviços externos, como `abacatepay.service.js` e `license.service.js`.

#### Frontend (Pasta `app/frontend/`)
O frontend é uma aplicação React que utiliza o Vite como ferramenta de build. Ele é responsável pela interface do usuário e pela interação com o backend.

- **index.html**: Arquivo HTML principal.
- **vite.config.js**: Configuração do Vite.
- **src/**: Contém o código-fonte principal do frontend.
  - **assets/**: Arquivos estáticos, como fontes e imagens.
  - **components/**: Componentes reutilizáveis, como `AppLayout.jsx` e `Sidebar.jsx`.
  - **context/**: Gerenciamento de estado global com o Context API, como `AppContext.jsx`.
  - **pages/**: Páginas principais da aplicação, como `CadastroPage.jsx` e `PixPaymentPage.jsx`.
  - **services/**: Serviços para comunicação com o backend, como `apiService.js`.

#### Comunicação Assíncrona
O projeto utiliza o RabbitMQ para gerenciar filas e processar tarefas de forma assíncrona. Por exemplo:
- Estratégias de criação de usuários e pagamentos Pix são processadas em filas específicas.
- O backend envia mensagens para as filas, e consumidores dedicados processam essas mensagens.

#### SSE (Server-Sent Events)
O frontend utiliza SSE para receber atualizações em tempo real do backend. Por exemplo:
- Após o cadastro de um usuário, uma conexão SSE é aberta para monitorar o status do pagamento Pix.

### Tecnologias Utilizadas
- **Backend**:
  - Node.js
  - RabbitMQ
  - Swagger para documentação da API
- **Frontend**:
  - React
  - Vite
  - Context API para gerenciamento de estado
  - Framer Motion para animações
- **Outros**:
  - Wails para integração com o sistema operacional

### Como Executar o Projeto

#### Backend
1. Navegue até a pasta `api/`.
2. Execute `docker-compose up` para iniciar os serviços necessários (ex.: RabbitMQ).
3. Instale as dependências com `npm install`.
4. Inicie o servidor com `npm start`.

#### Frontend
1. Navegue até a pasta `app/frontend/`.
2. Instale as dependências com `npm install`.
3. Inicie o servidor de desenvolvimento com `npm run dev`.

#### Aplicação Wails
1. Navegue até a pasta `app/`.
2. Execute `wails dev` para iniciar a aplicação desktop.

### Contribuição
Contribuições são bem-vindas! Sinta-se à vontade para abrir issues ou pull requests neste repositório.

### Licença
Este projeto está licenciado sob a licença MIT. Veja o arquivo LICENSE para mais detalhes.

## Diagrama de Fluxo

O diagrama abaixo representa o fluxo de execução do projeto SaaS, desde o cadastro do usuário até a ativação da licença:

```mermaid
flowchart TD
    CadastroPage -->|POST /api/users| Backend
    Backend -->|SSE /api/sse/[userId]| CadastroPage
    CadastroPage -->|Navigate| PixPaymentPage
    PixPaymentPage -->|Check License| Backend
    PixPaymentPage -->|Activate License| Dashboard

    Backend -->|Queue CREATE_USER_STRATEGY| UserStrategyConsumer
    UserStrategyConsumer -->|Queue CREATE_PIX_STRATEGY| PixStrategyConsumer
    PixStrategyConsumer -->|Process Payment| AbacatePayService
    AbacatePayService -->|Payment Status| PixStrategyConsumer
    PixStrategyConsumer -->|License Activated| Backend
```