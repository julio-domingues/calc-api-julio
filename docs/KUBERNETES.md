# Kubernetes existente

1-Pré-requisitos: Docker com buildx, Go para os testes e kubectl configurado para o
2-cluster. Execute na raiz do projeto.
3- O build e a publicação são feitos por voce  no registro que achar mais adequado (ECR, Artifactory ou outro).

## 1. Construir e publicar a imagem

Defina a referência do repositório e a plataforma dos nós:

```sh
export IMAGE=SEU_REGISTRO/calc-api-julio
export PLATFORM=linux/amd64
export TAG=v2
go test ./...
```

Substitua `SEU_REGISTRO` pelo endereço e caminho reais do registro.

**A plataforma deve corresponder aos nós do cluster, não ao computador do build.**
Confira a sua  arquitetura dos nós:

```sh
kubectl get nodes -L kubernetes.io/arch
```

- Nós `amd64` (Intel/AMD): mantenha `PLATFORM=linux/amd64`.
- Nós `arm64` (ARM): use `PLATFORM=linux/arm64`.

Uma imagem de arquitetura diferente pode falhar com `exec format error`.
Em clusters com ambas as arquiteturas, publique uma imagem multiarch ou restrinja
o agendamento aos nós compatíveis.

Faça a autenticação conforme o registro utilizado e publique:

```sh
docker buildx build --platform "$PLATFORM" -t "$IMAGE:$TAG" --push .
```

## 2. Fazer o deploy

Em `k8s/deployment.yaml`, substitua `image` pela referência publicada com a tag:

```yaml
image: SEU_REGISTRO/calc-api-julio:v2
```

Os nós precisam de acesso de leitura ao registro, pela identidade do provedor ou por
`imagePullSecrets` em `spec.template.spec`, conforme a configuração sua configuracao

```sh
kubectl create namespace calc-api-julio
kubectl apply -n calc-api-julio -f k8s/deployment.yaml -f k8s/service.yaml
kubectl rollout status -n calc-api-julio deployment/calc-api-julio --timeout=180s
kubectl get pods,svc -n calc-api-julio
```

Se o namespace já existir, pule sua criação. Aguarde as duas réplicas ficarem
`1/1 Running`. O Service é interno ao cluster, na porta 8000.

## 3. Testar

Execute uma chamada de dentro do cluster:

```sh
kubectl run teste -n calc-api-julio --rm -i --restart=Never --image=curlimages/curl:8.12.1 -- -fsS 'http://calc-api-julio:8000/api/sub?term_one=4&term_two=1'
```

Resposta esperada: `{"result":3}`.

Para abrir a calculadora no navegador:

```sh
kubectl port-forward -n calc-api-julio service/calc-api-julio 8082:8000
```

Abra http://localhost:8082/. Deixe o terminal aberto; Ctrl+C encerra o túnel.

## 4. Atualizar

Altere o código e publique uma imagem com uma tag nova:

```sh
go test ./...
export TAG=v3
docker buildx build --platform "$PLATFORM" -t "$IMAGE:$TAG" --push .
```

Troque o campo `image` em `k8s/deployment.yaml` para a nova tag:

```yaml
image: SEU_REGISTRO/calc-api-julio:v3
```

Aplique o manifesto para iniciar a atualização dos pods:

```sh
kubectl apply -n calc-api-julio -f k8s/deployment.yaml
kubectl rollout status -n calc-api-julio deployment/calc-api-julio --timeout=180s
```

Use uma tag nova por versão (`v2`, `v3`, `v4`...). Alterar a tag no manifesto
e executar `apply` inicia o rollout. Apenas publicar uma imagem no registro
não atualiza os pods automaticamente.

Para voltar à versão anterior, restaure a tag anterior no manifesto e aplique
novamente.

Repita o teste do passo 3. Se usar o navegador, reinicie o port-forward.

## 5. Remover

```sh
kubectl delete -n calc-api-julio -f k8s/deployment.yaml -f k8s/service.yaml
```

Se o namespace foi criado apenas para este projeto:

```sh
kubectl delete namespace calc-api-julio
```

## Se algo falhar

Confira os eventos dos pods e os logs da aplicação:

```sh
kubectl describe pods -n calc-api-julio -l app=calc-api-julio
kubectl logs -n calc-api-julio -l app=calc-api-julio --tail=50
```
