# ESA Horaro Proxy

Proxy for Horaro to allow CORS requests. It also slightly modifies the payload to more reasonable JSON and enables some caching for performance reasons.

## Build (outdated, just run `docker compose up`)

Go to the root folder and execute the following steps:

Install the dependencies

`$ go get -d ./...`

Build the application

`$ go install -v ./...`

Run the application

`$ go run src/*`

## Docker

To build the docker container, either:

`$ docker build -t esamaraton/esahoraroproxy`

or, with docker-compose:

`$ docker-compose build`

## Routes

**GET** `/v2/esa/schedule/{endpoint}`:

  Get all the speedruns for an event.

**GET** `/v2/esa/upcoming/{endpoint}?amount={int}`:

  Get the upcoming speedruns for an event (amount is optional, default is 5)

  `{endpoint}` can be in form:

  - `https://horaro.org/esa/2019-one.json`
  - `https://horaro.org/esa/2018-two`
  - `2018-one.json`
  - `2017-two`

## LICENSE

[MIT Copyright (c) 2019 European Speedrunner Assembly](./LICENSE)
