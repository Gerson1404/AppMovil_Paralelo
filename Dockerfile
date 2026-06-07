FROM golang:alpine AS builder
WORKDIR /app

ENV GOPROXY=https://proxy.golang.org,direct
COPY go.mod go.sum* ./
RUN go mod download
COPY . .

# LA MAGIA ESTÁ AQUÍ: CGO_ENABLED=0 obliga a Go a crear un binario 100% estático e independiente
RUN CGO_ENABLED=0 GOOS=linux go build -o apiserver ./cmd/main.go

FROM alpine:latest
WORKDIR /root/
# Copiamos el binario estático desde la fase de compilación
COPY --from=builder /app/apiserver .
RUN mkdir -p uploads 
EXPOSE 8080
CMD ["./apiserver"]