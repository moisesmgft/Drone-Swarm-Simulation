FROM golang:1.20
WORKDIR /app
COPY ./go/ /app
RUN go mod init project && go mod tidy
CMD ["go", "run", "drone/drone.go"]