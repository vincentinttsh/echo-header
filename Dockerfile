# build stage
FROM golang:1.17-alpine AS build-env
COPY . /src
RUN apk add build-base && cd /src && go build -ldflags "-s -w" -o app

# final stage
FROM alpine:3
WORKDIR /app
COPY --from=build-env /src/app /app/app
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser
ENTRYPOINT ["./app"]
EXPOSE 8080
ENV GIN_MODE=release