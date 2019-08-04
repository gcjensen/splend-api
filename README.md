# splend-api
The RESTful API behind the Splend (split & lend) app, a tool for managing
shared expenditure with your other half.

[![CircleCI](https://circleci.com/gh/gcjensen/splend-api/tree/master.svg?style=svg&circle-token=cd1f6a0dfb674a6e51208a65872cf8bb96bef46f)](https://circleci.com/gh/gcjensen/splend-api/tree/master)

This project is still under active development. Please see the project
[issues](https://github.com/gcjensen/splend-api/issues) for planned features
etc.

### Build & Development Instructions

- Install: `go get github.com/gcjensen/splend-api`
- `cd` into the `splend-api` directory
- Install the dependencies: `go get ./...`
- Create a database import the schema in `meta/schema.sql`
- Copy the `etc/splend-api.yaml` config file to `/etc/splend-api.yaml` and
  update it to reflect the details of your database created above
- Change to the main application directory: `cd cmd/splend-api/`
- Compile the app: `go build`
- Run the executable: `./splend-api`

### Running the tests

- Install the assertion library: `go get github.com/stretchr/testify/assert`
- Run all the tests including those in sub packages: `go test ./...`
