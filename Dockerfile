FROM golang:1.16-alpine

WORKDIR /app

COPY api-service/go.mod ./
COPY api-service/go.sum ./
RUN go mod download

COPY api-service/api.go ./

RUN go build -o api api.go

EXPOSE 8080
EXPOSE 2121

CMD [ "./api" ]