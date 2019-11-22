# ESA Horaro Proxy

Proxy for Horaro to allow CORS requests. It also slightly modifies the payload to more reasonable JSON and enables some caching for performance reasons.

## Build

Go to the root folder and execute the following steps:

Install the dependencies

`$ go get -d ./...`

Build the application

`$ go install -v ./...`

## Docker

To build the docker container, either:

`$ docker build -t esamaraton/esahoraroproxy`

or, with docker-compose:

`$ docker-compose build`

## Routes

GET `/v1/esa/{endpoint}`: Get the events for a specific year

Where endpoint can be in form:

- `https://horaro.org/esa/2019-one.json`
- `https://horaro.org/esa/2018-two`
- `2018-one.json`
- `2017-two`

## LICENSE

[MIT Copyright (c) 2019 European Speedrunner Assembly](./LICENSE)
