FROM docker.io/golang:1.22.12-alpine3.21 AS compiler
WORKDIR /app/
COPY go.mod .
COPY main.go .
RUN go build -o main .

FROM scratch AS runner
COPY --from=compiler /app/main /main
USER 1001:1001
ENTRYPOINT ["/main"]
EXPOSE 8080/tcp