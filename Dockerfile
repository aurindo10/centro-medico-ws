# Utilize uma imagem oficial de Go para compilar o código
FROM golang:1.20 AS build

# Define o diretório de trabalho dentro do container
WORKDIR /server

# Copia os arquivos go.mod e go.sum e instala as dependências
COPY go.mod go.sum ./
RUN go mod download

# Copia os arquivos do código fonte para o container
COPY . .

# Compila o código. Isso resultará em um binário chamado "server" no diretório /app.
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd

# Utilize uma imagem mais leve para rodar o binário
FROM alpine:latest

# Copia o binário compilado da imagem build para a imagem atual
COPY --from=build /server .

# Porta que o servidor escutará
EXPOSE 8080

# Comando para rodar o binário
CMD ["/server"]
