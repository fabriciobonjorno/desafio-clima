# Clima por CEP com OpenTelemetry e Zipkin

Aplicação Go com dois serviços HTTP. O Serviço A valida o CEP e encaminha o pedido ao Serviço B, que consulta o ViaCEP e o Open-Meteo, calcula as temperaturas e devolve a cidade. Ambos exportam traces via OTLP para o OpenTelemetry Collector, que encaminha os dados ao Zipkin.

## Subir o ambiente

Requer Docker com Docker Compose.

```bash
docker compose up --build
```

O Compose sobe Serviço A, Serviço B, OpenTelemetry Collector e Zipkin. O Serviço A fica disponível em `http://localhost:8080`; a interface do Zipkin fica em `http://localhost:9411`.

Se a porta local `8080` já estiver ocupada, defina outra porta para o Serviço A, por exemplo `SERVICE_A_PORT=18080 docker compose up --build` e use `http://localhost:18080`.

## Requisição

Envie um `POST` para o Serviço A:

```bash
curl -i http://localhost:8080/ \
  -H 'Content-Type: application/json' \
  -d '{"cep":"01001000"}'
```

Sucesso (`200 OK`):

```json
{
  "city": "São Paulo",
  "temp_C": 28.5,
  "temp_F": 83.3,
  "temp_K": 301.5
}
```

CEP inválido ou corpo JSON malformado retorna `422` com `invalid zipcode`. Um CEP válido, mas não encontrado no ViaCEP, retorna `404` com `can not find zipcode`. Falhas nos provedores externos retornam `502`.

As conversões seguem as fórmulas do enunciado: `F = C × 1.8 + 32` e `K = C + 273`. Isso produz `301.5 K` para `28.5 °C`; o exemplo do desafio mostra `301.65 K`, valor correspondente a `C + 273.15`.

## Ver os traces no Zipkin

1. Faça a requisição de exemplo acima.
2. Abra [http://localhost:9411](http://localhost:9411).
3. Selecione o serviço `service-a` ou `service-b` e clique em **Run Query**.
4. Abra o trace para ver a sequência `service-a → service-b`, além dos spans manuais `zipcode.lookup` e `temperature.lookup` no Serviço B.

O exportador agrupa spans antes de enviá-los; após a requisição, aguarde alguns segundos para o trace aparecer.

O contexto W3C Trace Context é extraído pelo servidor A e injetado no pedido HTTP para B. Os serviços enviam spans em OTLP/HTTP ao Collector; o Collector exporta os traces ao Zipkin.

## Testes

```bash
go test -race ./...
```

Os testes usam `httptest` para simular ViaCEP, Open-Meteo e o Serviço B. O teste de fluxo distribuído verifica que os spans de A, B e das consultas externas compartilham o mesmo trace ID.

Para parar os containers:

```bash
docker compose down
```
