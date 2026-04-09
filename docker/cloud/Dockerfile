FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY ./bin /app
COPY ./docs/openapi.yaml /app/docs/openapi.yaml

# Default entrypoint (app)
ENTRYPOINT ["/app/app"]