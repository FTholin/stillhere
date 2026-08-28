FROM golang:1.26-alpine AS build
WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/stillhere ./cmd/stillhere

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/stillhere /stillhere
EXPOSE 8080
ENTRYPOINT ["/stillhere"]