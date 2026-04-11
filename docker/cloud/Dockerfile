FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY ./bin /app
COPY ./docs/ /app/docs/

# Default entrypoint (app)
ENTRYPOINT ["/app/app"]