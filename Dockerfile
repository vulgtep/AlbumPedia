
FROM golang:alpine

WORKDIR /app

COPY . .

RUN go build -o mi-servidor main.go

EXPOSE 8080

CMD ["/app/mi-servidor"]