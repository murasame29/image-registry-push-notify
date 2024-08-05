FROM golang:1.22.2 AS build
WORKDIR /build

COPY . .

RUN go mod download \
&& go build -o main cmd/main.go

FROM gcr.io/distroless/static-debian11:nonroot
WORKDIR /app

COPY --from=build /build/main /app/main

CMD ["/app/main"]