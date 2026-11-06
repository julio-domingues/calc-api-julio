# calc-api-julio

Calculadora em Go com interface web e uma API com as quatro operações básicas.

## Baixar e executar

Você precisa de Git e Go 1.24 ou superior:

```sh
git clone https://github.com/julio-domingues/calc-api-julio.git
cd calc-api-julio
go run .
```

Abra [localhost:8000](http://localhost:8000/), informe os números e clique em
**Calcular**. O link **Ver resposta JSON** abre a resposta da API em outra aba.
Para encerrar, pressione Ctrl+C no terminal.

## Testar a API

Com a aplicação rodando, execute em outro terminal:

```sh
curl 'http://localhost:8000/api/sub?term_one=4&term_two=1'
```

Resposta esperada:

```json
{"result":3}
```

Todos os endpoints abaixo usam GET e recebem os inteiros `term_one` e `term_two`:

| Endpoint | Operação |
| --- | --- |
| `/api/sum` | Soma |
| `/api/sub` | Subtração |
| `/api/mul` | Multiplicação |
| `/api/div` | Divisão |

A divisão descarta a parte decimal: `5 / 2 = 2` e `-5 / 2 = -2`.
O resultado da divisão é inteiro, com truncamento em direção a zero.
Divisão por zero retorna **400**; parâmetros ausentes ou inválidos retornam **422**.
Os health checks ficam em `/healthz` e `/readyz`.

## Rodar os testes

```sh
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

Abra `coverage.html` no navegador para ver a cobertura. A última execução teve
78,5% de cobertura total e 100% nas funções HTTP; a inicialização do servidor
não é coberta pelos testes automatizados.

## Rodar com Docker

Com o Docker instalado e rodando, execute na pasta do projeto:

```sh
docker build -t calc-api-julio:v2 .
docker run --rm --name calc-api-julio -p 8081:8000 calc-api-julio:v2
```

Abra [localhost:8081](http://localhost:8081/). Para encerrar e remover o container,
use Ctrl+C no terminal. A aplicação roda como usuário sem privilégios.

## Kubernetes

- [Deploy em um cluster existente](docs/KUBERNETES.md): construir e publicar a imagem no registro da equipe,
  aplicar os manifests, testar, atualizar e remover a aplicação.
- [Teste local com kind](docs/KIND.md): criar um cluster local e fazer o deploy.

O Service usa ClusterIP na porta 8000. O monitoramento Prometheus é opcional e
ainda depende da implementação de `/metrics`.
