# Desafios de Clima por CEP

Este repositório reúne dois projetos independentes em Go para consultar o clima a partir de um CEP.

## Projetos

- [`desafio-basico/`](desafio-basico/README.md): API de consulta de clima por CEP, com testes, Docker e instruções de deploy no Google Cloud Run.
- [`desafio-observabilidade/`](desafio-observabilidade/README.md): Serviços A e B com tracing distribuído usando OpenTelemetry Collector e Zipkin.

## Executar o desafio básico

```bash
cd desafio-basico
go test ./...
go run ./cmd/server
```

A API fica em `http://localhost:8080`. Exemplo:

```bash
curl 'http://localhost:8080/?cep=01001000'
```

Para executar via Docker, consulte [`desafio-basico/README.md`](desafio-basico/README.md).

## Executar o desafio com observabilidade

Na raiz do repositório:

```bash
docker compose -f desafio-observabilidade/docker-compose.yaml up --build
```

O Serviço A fica disponível em `http://localhost:8080` e o Zipkin em `http://localhost:9411`. Para testar o segundo projeto sem Docker:

```bash
cd desafio-observabilidade
go test ./...
```

Consulte [`desafio-observabilidade/README.md`](desafio-observabilidade/README.md) para o payload POST e os passos para visualizar os traces.

## Repositório

https://github.com/fabriciobonjorno/desafio-clima
