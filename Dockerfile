FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod .
COPY main.go .
COPY web ./web
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /api .

FROM scratch
COPY --from=build /api /api
USER 10001:10001
EXPOSE 8000
ENTRYPOINT ["/api"]
