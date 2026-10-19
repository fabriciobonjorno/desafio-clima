# Clima por CEP

API em Go que recebe um CEP brasileiro, consulta a cidade no ViaCEP e obtém a temperatura atual pelo Open-Meteo.

## Acesso

- **Cloud Run:** ainda não implantado. Preencha aqui a URL pública depois do deploy.
- **GitHub:** ainda não publicado. Este workspace não possui um remoto configurado.

## API

```http
GET /?cep=01001000
```

Sucesso (`200`):

```json
{"temp_C":28.5,"temp_F":83.3,"temp_K":301.5}
```

CEP malformado retorna `422` e `invalid zipcode`; CEP não encontrado retorna `404` e `can not find zipcode`. Falhas nas APIs externas retornam `502`.

As temperaturas são arredondadas para uma casa decimal em Celsius e Fahrenheit e duas em Kelvin. A conversão Kelvin segue a fórmula deste desafio (`C + 273`).

**Nota sobre o contrato:** o exemplo informado associa `28.5 °C` a `301.65 K`, valor que corresponde a `C + 273.15`. A fórmula escrita no desafio (`C + 273`) resulta em `301.5 K`; a implementação segue a fórmula escrita.

## Rodar localmente

Requer Go 1.23 ou superior.

```bash
go run ./cmd/server
curl 'http://localhost:8080/?cep=01001000'
```

## Testes

```bash
go test ./...
```

Os testes de integração simulam ViaCEP e Open-Meteo localmente e não dependem da rede.

## Docker

```bash
docker build -t clima-por-cep .
docker run --rm -p 8080:8080 clima-por-cep
curl 'http://localhost:8080/?cep=01001000'
```

## Deploy no Google Cloud Run

Com `gcloud` instalado e autenticado, selecione um projeto Google Cloud com faturamento habilitado e execute:

```bash
gcloud config set project SEU_PROJECT_ID
gcloud run deploy clima-por-cep \
  --source . \
  --region southamerica-east1 \
  --allow-unauthenticated \
  --port 8080
```

Após o deploy, copie a URL exibida pelo comando para a seção **Acesso** acima. A aplicação não contém chaves de API; ViaCEP e Open-Meteo são chamados diretamente.
