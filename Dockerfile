FROM golang:1.26

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o scheduler ./main.go

EXPOSE 7540

CMD ["./scheduler"]