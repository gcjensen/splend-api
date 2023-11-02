# splend-api
The RESTful API behind the Splend (split & lend) app, a tool for managing
shared expenditure with your other half.

[![CircleCI](https://dl.circleci.com/status-badge/img/gh/gcjensen/splend-api/tree/master.svg?style=svg)](https://dl.circleci.com/status-badge/redirect/gh/gcjensen/splend-api/tree/master)

### Build Instructions

The API, nginx reverse proxy and mysql database can all be brought up using
docker compose:

```
docker-compose -f docker-compose.yaml -f dev.yaml up
```

Bare in mind you'll need to import `meta/schema.sql` on the initial run of the
mysql container.

### Running the tests

- Install the assertion library: `go get github.com/stretchr/testify/assert`
- Run all the tests including those in sub packages: `go test ./...`

#### CI

CircleCI is used to build the project and run the tests when code is pushed.
This can be tested locally using the
[CircleCI Local CLI](https://circleci.com/docs/2.0/local-cli/). Once
[installed and setup](https://circleci.com/docs/2.0/local-cli/#installation),
the config in
[.circleci/config.yml](https://github.com/gcjensen/splend-api/blob/master/.circleci/config.yml)
can be validated with:

```
circleci config validate
```

The CI job can also be fully run locally inside a Docker container. Once Docker
has been [installed and setup](https://docs.docker.com/install/), this can be
done with:

```
circleci local execute --job build
```

### Monzo webhooks

The app supports automatically adding transactions from a Monzo account.
To register a Monzo webhook, you must first retrieve an account ID and access
token for the user. Please refer to
[Monzo's documentation]([https://docs.monzo.com/#introduction](https://docs.monzo.com/#introduction))
for instructions on doing so. Once you have these, you can register the webhook
with the following POST request (this example uses
[httpie](https://httpie.org/)):

```
    http --form POST "https://api.monzo.com/webhooks" \
        "Authorization: Bearer $access_token" \
        "account_id=$account_id" \
        "url=$url_of_your_server:3002/user/$user_id/monzo-webhook"
```

Please refer to Monzo's documentation for instructions on deleting registered
webhooks.
