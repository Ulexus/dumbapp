FROM golang:latest AS build

ENV CGO_ENABLED=0

COPY . .
RUN go build -o /app

FROM scratch
COPY --from=build /app /app

ENTRYPOINT ["/app"]
CMD []
