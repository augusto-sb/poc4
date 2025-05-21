#docker container run --detach --rm -p "6379:6379/tcp" --tmpfs "/data" redis:7.4.3-alpine
REDIS_URI="redis://localhost:6379/0" LOGGER="true" CORS_ORIGIN="http://localhost:4200" go run .