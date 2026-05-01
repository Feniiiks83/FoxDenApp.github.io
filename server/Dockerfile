FROM golang:latest

WORKDIR /server

COPY . .

RUN chmod +x /server/parser/Parser

EXPOSE 3001

CMD ["go", "run", "./cmd/tax-parser/main.go"]