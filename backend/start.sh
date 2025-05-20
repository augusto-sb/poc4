docker container run --detach --rm -p "6379:6379/tcp" --tmpfs "/data" redis:7.4.3-alpine
LOGGER="true" CORS_ORIGIN="http://localhost:4200" go run .