FROM golang:1.22.2 AS build
WORKDIR /build

COPY . .

RUN go mod download \
&& go build -o main cmd/main.go

FROM public.ecr.aws/lambda/provided:al2023
WORKDIR /app

COPY --from=build /build/main /app/main

USER 1001

ENTRYPOINT ["/app/main"]