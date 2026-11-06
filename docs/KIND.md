# Kubernetes local com kind

Você precisa de Docker rodando, kind e kubectl. Execute na pasta do projeto.

## 1. Subir a aplicação

```sh
docker build -t calc-api-julio:v2 .
kind create cluster --name calc-api-julio
kubectl config use-context kind-calc-api-julio
kind load docker-image calc-api-julio:v2 --name calc-api-julio
kubectl create namespace calc-api-julio
kubectl apply -n calc-api-julio -f k8s/deployment.yaml -f k8s/service.yaml
kubectl rollout status -n calc-api-julio deployment/calc-api-julio --timeout=180s
```

Este guia usa `image: calc-api-julio:v2` no Deployment. Se alterou esse campo
para um registro externo, volte para esse valor antes de aplicar.

## 2. Testar

De dentro do cluster:

```sh
kubectl run teste -n calc-api-julio --rm -i --restart=Never --image=curlimages/curl:8.12.1 -- -fsS 'http://calc-api-julio:8000/api/sub?term_one=4&term_two=1'
```

Resposta esperada: `{"result":3}`.

Para abrir no navegador:

```sh
kubectl port-forward -n calc-api-julio service/calc-api-julio 8082:8000
```

Abra http://localhost:8082/. Ctrl+C encerra o túnel.

## 3. Atualizar

Depois de alterar o código:

```sh
go test ./...
docker build -t calc-api-julio:v3 .
kind load docker-image calc-api-julio:v3 --name calc-api-julio
```

Em `k8s/deployment.yaml`, troque a tag da imagem para `v3` e execute:

```sh
kubectl apply -n calc-api-julio -f k8s/deployment.yaml
kubectl rollout status -n calc-api-julio deployment/calc-api-julio --timeout=180s
```

Repita o teste do passo 2. Reinicie o port-forward se estiver usando o navegador.

## 4. Remover

```sh
kubectl delete -n calc-api-julio -f k8s/deployment.yaml -f k8s/service.yaml
kubectl delete namespace calc-api-julio
kind delete cluster --name calc-api-julio
```
