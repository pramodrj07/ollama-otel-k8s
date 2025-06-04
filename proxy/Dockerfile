FROM golang:1.24-alpine as builder

WORKDIR /app
COPY . .

RUN go mod tidy 
RUN go build -o proxy .
EXPOSE 8080

CMD ["./proxy"]