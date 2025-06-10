# ⚠️ REPOSITORY ARCHIVED ⚠️

> **IMPORTANT NOTICE:** This repository has been archived and moved to the official AstroVista organization on GitHub.
>
> **New Organization:** [https://github.com/AstroVista](https://github.com/AstroVista/astrovista-api)
>
> Please go to the new repository for the latest version, to submit issues, or to contribute to the project.

---

# AstroVista API

API para gerenciar dados da NASA APOD (Astronomy Picture of the Day) com recursos avançados de documentação interativa, sistema de cache e suporte a múltiplos idiomas.

## Recursos

-   **Documentação interativa** com Swagger/OpenAPI
-   **Sistema de cache** com Redis para melhor performance
-   **Internacionalização (i18n)** com suporte a múltiplos idiomas - Inglês (padrão)
    -   Português do Brasil
    -   Espanhol
    -   Francês
    -   Alemão
    -   Italiano
    -   Suporte para adicionar novos idiomas facilmente

## Configuração

### Requisitos

-   Go 1.18 ou superior
-   MongoDB (para armazenamento de dados)
-   Redis (opcional, para cache)

### Variáveis de ambiente

-   `PORT` - Porta do servidor (padrão: 8081)
-   `MONGODB_URI` - URI de conexão com o MongoDB
-   `REDIS_URL` - URL do servidor Redis (opcional)
-   `REDIS_PASSWORD` - Senha do Redis (opcional)
-   `GOOGLE_TRANSLATE_API_KEY` - Chave da API do Google Translate (opcional)
-   `DEEPL_API_KEY` - Chave da API DeepL para traduções (opcional)

## Serviços de Tradução

A API pode utilizar diferentes serviços de tradução:

### Google Translate

Para usar o Google Translate, você precisa:

1. Criar uma conta no [Google Cloud Platform](https://cloud.google.com/)
2. Criar um novo projeto
3. Ativar a Cloud Translation API
4. Criar uma chave de API
5. Definir a variável de ambiente `GOOGLE_TRANSLATE_API_KEY`

```bash
export GOOGLE_TRANSLATE_API_KEY="sua-chave-aqui"
```

### DeepL

Para usar o DeepL, você precisa:

1. Criar uma conta no [DeepL API](https://www.deepl.com/pro-api)
2. Obter sua chave de autenticação
3. Definir a variável de ambiente `DEEPL_API_KEY`

```bash
export DEEPL_API_KEY="sua-chave-aqui"
```

### Simulação (Mock)

Se nenhuma chave de API for configurada, a API usará um serviço de tradução simulado para desenvolvimento.

## Sistema de Cache de Traduções

Para melhorar a performance e evitar requisições repetidas às APIs de tradução, implementamos um sistema de cache em dois níveis:

1. **Cache em memória**: Armazena traduções recentes na memória para acesso rápido
2. **Cache Redis**: Se o Redis estiver disponível, as traduções também são armazenadas de forma persistente

As traduções são armazenadas por 30 dias no cache Redis, reduzindo significativamente o número de chamadas às APIs externas.

## Executando a API

```bash
# Clone o repositório
git clone https://github.com/seu-usuario/astrovista-api.git
cd astrovista-api

# Instale as dependências
go get -u

# Execute a API
go run main.go
```

## Endpoints

### Documentação

-   `/swagger/` - Documentação interativa Swagger

### Principais endpoints

-   `GET /apod` - Obtém o APOD mais recente
-   `GET /apod/{date}` - Obtém um APOD por data específica
-   `GET /apods` - Lista todos os APODs cadastrados
-   `GET /apods/search` - Pesquisa avançada com filtros
-   `GET /apods/date-range` - Busca APODs por intervalo de datas
-   `GET /languages` - Lista idiomas suportados
-   `POST /apod` - Adiciona um novo APOD

## Suporte a idiomas

Para obter respostas em um idioma específico, você pode:

1. Enviar o cabeçalho `Accept-Language` na requisição

    ```
    Accept-Language: pt-BR
    ```

2. Ou adicionar o parâmetro `lang` na URL
    ```
    /apod?lang=pt-BR
    ```

## Cache

As respostas da API incluem o cabeçalho `X-Cache` para indicar se o resultado veio do cache:

-   `X-Cache: HIT` - Resposta recuperada do cache
-   `X-Cache: MISS` - Resposta obtida do banco de dados

## Licença

MIT

---

## About the archived repository

This repository was archived on June 10, 2025, and moved to the official AstroVista organization on GitHub to centralize development and improve collaboration among contributors. All issues, pull requests, and discussions should be directed to the [new repository](https://github.com/AstroVista/astrovista-api).

**No new contributions will be accepted in this repository.**
