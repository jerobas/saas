import swaggerUi from 'swagger-ui-express';

const swaggerSpec = {
  openapi: '3.0.0',
  info: {
    title: 'SaaS Licensing API',
    description: 'Backend para sistema de licenciamento com pagamento PIX via AbacatePay',
    version: '1.0.0',
    contact: {
      name: 'Development Team',
    },
  },
  servers: [
    {
      url: 'http://localhost:3000',
      description: 'Development server',
    },
    {
      url: 'https://api.vezono.com/saas/api',
      description: 'Production server',
    },
  ],
  paths: {
    '/api/create': {
      post: {
        tags: ['Users'],
        summary: 'Criar usuário e gerar PIX',
        description: 'Cria um novo usuário, registra no AbacatePay e gera PIX para pagamento da licença',
        requestBody: {
          required: true,
          content: {
            'application/json': {
              schema: {
                type: 'object',
                properties: {
                  email: { type: 'string', format: 'email', example: 'user@example.com' },
                  name: { type: 'string', example: 'João Silva' },
                  taxId: { type: 'string', example: '12345678901', description: 'CPF ou CNPJ' },
                  cellphone: { type: 'string', example: '11999999999' },
                },
                required: ['email', 'name', 'taxId', 'cellphone'],
              },
            },
          },
        },
        responses: {
          201: {
            description: 'Usuário criado e PIX gerado com sucesso',
            content: {
              'application/json': {
                schema: {
                  type: 'object',
                  properties: {
                    success: { type: 'boolean' },
                    data: {
                      type: 'object',
                      properties: {
                        userId: { type: 'string', format: 'uuid' },
                        paymentId: { type: 'string', format: 'uuid' },
                        brCode: { type: 'string', description: 'Código PIX copiar e colar' },
                        brCodeBase64: { type: 'string', description: 'QR Code em base64' },
                        expiresAt: { type: 'string', format: 'date-time' },
                        amount: { type: 'number', example: 500 },
                      },
                    },
                  },
                },
              },
            },
          },
          400: { description: 'Dados inválidos ou email já cadastrado' },
          500: { description: 'Erro ao criar usuário ou PIX' },
        },
      },
    },
    '/api/renew': {
      post: {
        tags: ['Subscriptions'],
        summary: 'Renovar assinatura',
        description: 'Gera novo PIX para usuário existente renovar sua licença',
        requestBody: {
          required: true,
          content: {
            'application/json': {
              schema: {
                type: 'object',
                properties: {
                  userId: { type: 'string', format: 'uuid', description: 'ID do usuário (alternativa: email)' },
                  email: { type: 'string', format: 'email', description: 'Email do usuário (alternativa: userId)' },
                },
              },
            },
          },
        },
        responses: {
          200: {
            description: 'PIX gerado com sucesso',
            content: {
              'application/json': {
                schema: {
                  type: 'object',
                  properties: {
                    success: { type: 'boolean' },
                    data: {
                      type: 'object',
                      properties: {
                        userId: { type: 'string', format: 'uuid' },
                        paymentId: { type: 'string', format: 'uuid' },
                        brCode: { type: 'string' },
                        brCodeBase64: { type: 'string' },
                        expiresAt: { type: 'string', format: 'date-time' },
                        amount: { type: 'number' },
                      },
                    },
                  },
                },
              },
            },
          },
          400: { description: 'Usuário não encontrado ou userId/email não fornecido' },
          500: { description: 'Erro ao gerar PIX' },
        },
      },
    },
    '/api/webhooks/payment': {
      post: {
        tags: ['Webhooks'],
        summary: 'Webhook de pagamento AbacatePay',
        description: 'Recebe notificação de pagamento, gera token de licença e envia por email',
        requestBody: {
          required: true,
          content: {
            'application/json': {
              schema: {
                type: 'object',
                properties: {
                  event: { type: 'string', enum: ['payment.confirmed', 'payment.cancelled', 'payment.expired'] },
                  data: {
                    type: 'object',
                    properties: {
                      id: { type: 'string', description: 'ID da transação PIX' },
                      status: { type: 'string' },
                    },
                  },
                },
              },
            },
          },
        },
        responses: {
          200: {
            description: 'Webhook processado com sucesso',
            content: {
              'application/json': {
                schema: {
                  type: 'object',
                  properties: {
                    success: { type: 'boolean' },
                    data: {
                      type: 'object',
                      properties: {
                        event: { type: 'string' },
                        paymentId: { type: 'string' },
                        status: { type: 'string' },
                      },
                    },
                  },
                },
              },
            },
          },
          401: { description: 'Assinatura de webhook inválida' },
          500: { description: 'Erro ao processar webhook' },
        },
      },
    },
    '/api/dev/simulate': {
      post: {
        tags: ['Development'],
        summary: 'Simular pagamento (DEV ONLY)',
        description: 'Simula um pagamento PIX completado. Apenas disponível em NODE_ENV=development',
        requestBody: {
          required: true,
          content: {
            'application/json': {
              schema: {
                type: 'object',
                properties: {
                  paymentId: { type: 'string', format: 'uuid' },
                },
                required: ['paymentId'],
              },
            },
          },
        },
        responses: {
          200: {
            description: 'Pagamento simulado com sucesso',
            content: {
              'application/json': {
                schema: {
                  type: 'object',
                  properties: {
                    success: { type: 'boolean' },
                    data: {
                      type: 'object',
                      properties: {
                        paymentId: { type: 'string' },
                        status: { type: 'string', example: 'PAID' },
                        message: { type: 'string' },
                        tokenSent: { type: 'boolean' },
                      },
                    },
                  },
                },
              },
            },
          },
          400: { description: 'Pagamento inválido ou já confirmado' },
          403: { description: 'Simulação só permitida em desenvolvimento' },
          404: { description: 'Pagamento não encontrado' },
        },
      },
    },
    '/api/license': {
      get: {
        tags: ['License'],
        summary: 'Verificar status da licença',
        description: 'Retorna o status da licença do usuário, incluindo token se ativa. Parâmetro: userId ou email',
        parameters: [
          {
            name: 'userId',
            in: 'query',
            schema: { type: 'string', format: 'uuid' },
            description: 'ID do usuário (alternativa: email)',
          },
          {
            name: 'email',
            in: 'query',
            schema: { type: 'string', format: 'email' },
            description: 'Email do usuário (alternativa: userId)',
          },
        ],
        responses: {
          200: {
            description: 'Status da licença retornado com sucesso',
            content: {
              'application/json': {
                schema: {
                  type: 'object',
                  properties: {
                    success: { type: 'boolean' },
                    data: {
                      type: 'object',
                      properties: {
                        userId: { type: 'string', format: 'uuid' },
                        email: { type: 'string', format: 'email' },
                        licenseActive: { type: 'boolean', example: true },
                        licenseExpiresAt: { type: 'string', format: 'date-time' },
                        licenseToken: { type: 'string', description: 'Token de licença assinado' },
                        daysRemaining: { type: 'integer', example: 365 },
                      },
                    },
                  },
                },
              },
            },
          },
          400: { description: 'userId ou email não fornecido' },
          404: { description: 'Usuário não encontrado' },
        },
      },
    },
    '/auth/check-license': {
      post: {
        tags: ['Authentication'],
        summary: 'Verificar se o usuário tem licença ativa',
        description: 'Verifica se o usuário tem licença ativa',
        requestBody: {
          required: true,
          content: {
            'application/json': {
              schema: {
                type: 'object',
                properties: {
                  email: { type: 'string', format: 'email', description: 'Email do usuário' },
                  password: { type: 'string', description: 'Senha do usuário para autenticação' },
                },
              },
            },
          },
        },
        responses: {
          200: {
            description: 'Usuário tem licença ativa',
            content: {
              'application/json': {
                schema: {
                  type: 'object',
                  properties: {
                    success: { type: 'boolean' },
                    data: {
                      type: 'object',
                      properties: {
                        userId: { type: 'string', format: 'uuid' },
                        email: { type: 'string', format: 'email' },
                        licenseActive: { type: 'boolean', example: true },
                        licenseExpiresAt: { type: 'string', format: 'date-time' },
                        licenseToken: { type: 'string', description: 'Token de licença assinado' },
                        daysRemaining: { type: 'integer', example: 365 },
                      },
                    },
                  },
                },
              },
            },
          },
          400: { description: 'Dados inválidos ou email já cadastrado' },
          500: { description: 'Erro ao verificar se o usuário tem licença ativa' },
        },
      },
    },
    '/health': {
      get: {
        tags: ['Health'],
        summary: 'Verificar status da saúde do servidor',
        description: 'Retorna o status da saúde do servidor',
        responses: {
          200: {
            description: 'Servidor está funcionando',
            content: {
              'application/json': {
                schema: {
                  type: 'object',
                  properties: {
                    status: { type: 'string', example: 'OK' },
                    timestamp: { type: 'string', example: '2023-04-05T12:34:56.789Z' },
                  },
                },
              },
            },
          },
        },
      },
    },
  },
  components: {
    schemas: {
      User: {
        type: 'object',
        properties: {
          id: { type: 'string', format: 'uuid' },
          email: { type: 'string', format: 'email' },
          name: { type: 'string' },
          taxId: { type: 'string' },
          cellphone: { type: 'string' },
          abacatePayCustomerId: { type: 'string' },
          createdAt: { type: 'string', format: 'date-time' },
          updatedAt: { type: 'string', format: 'date-time' },
        },
      },
      Payment: {
        type: 'object',
        properties: {
          id: { type: 'string', format: 'uuid' },
          userId: { type: 'string', format: 'uuid' },
          abacatePayCustomerId: { type: 'string' },
          abacatePayPixId: { type: 'string' },
          amount: { type: 'number', description: 'Valor em centavos' },
          status: { type: 'string', enum: ['PENDING', 'PAID', 'CANCELLED', 'EXPIRED'] },
          brCode: { type: 'string' },
          brCodeBase64: { type: 'string' },
          expiresAt: { type: 'string', format: 'date-time' },
          error: { type: 'string' },
          createdAt: { type: 'string', format: 'date-time' },
          updatedAt: { type: 'string', format: 'date-time' },
        },
      },
    },
  },
};

export const setupSwagger = (app) => {
  app.use('/api/api-docs', swaggerUi.serve, swaggerUi.setup(swaggerSpec));
};
