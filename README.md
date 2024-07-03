# Wallet Service

## Tech stack
- Postgres for the database
- Nats (Jetstream) as the event bus
- HTTP for the API
- Docker
- golangci-lint for the linter 
- go-fumpt for stricter formatting

## API 
- POST /v1/wallet - creates a wallet
- GET /v1/wallet/:walletID - gets a wallet by id it also includes balance, pending credit and debit balance 
- POST /v1/wallet/:walletID/transfers/credit - credit in this case means removing money from the wallet (the term is taken from accounting)
- POST /v1/wallet/:walletID/transfers/debit - debit in this case means adding money to the wallet (the term is taken from accounting)
- POST /v1/wallet/:walletID/transfers/:transferID/complete - completes a transfer, this is a separate step to allow for rolling back a transfer, but requires the debit/credit to be in a pending state initially 
- POST /v1/wallet/:walletID/transfers/:transferID/revert - rolls back (marks it as failed in the projection) a transfer, this is a separate step to allow for rolling back a transfer, but requires the debit/credit to be in a pending state initially

## Structure
- cmd/ - contains the main package (entry point for the service) this includes both the api and worker commands so a single binary can run both
  - api/ - contains the http server and the routes
  - worker/ - contains the event worker that listens for events and updates the wallet state
- migrations/ - contains database migrations and migration state (for atlas) 
- specification/ - contains a postman collection and a basic openapi spec
- internal/
    - api/ - contains entity/domain types and business logic
        - app 
           - contract - contains all interfaces that are used in the app
           - entity - contains all domain types 
           - request - contains all request types
           - response - contains all response types
        - wallet - contains all wallet related business logic/implementations
            - tests - contains integration/dockertest tests 
    - pkg - contains packages and patterns that I typically use in my projects (those were not built for this project)
## Postman collection
A postman collection and a basic openapi spec is included in the specification dir

## Starting the service
- `make up` will start the service, after that you should be able to access the service on `localhost:8089`

## Decisions 
- Projections are stored in a separate table instead of a separate db/datastore, again this is for simplicity sake and could easily be changed
- The service by default doesn't require an idempotency key (a randomly generated one is used), but it can be provided with the caveat that it has to be unique in the context of all transactions in the given wallet
- UUIDv7 is used across the board as it provides a k-sortable time based UUID, which is useful for sorting/ordering and comparisons
- Atlas is used for database migrations so the migrations/development process has a few extra steps 
- I've copied over a couple of packages I typically use in my personal projects, they typically live in a monorepo, but for simplicity sake they are copied over here (excluding tests)
- Dockertest is used for "integration" postgres/nats tests 
- Events are only emitted on transfer routes (no wallet created event)
- Wallet state/projections are updated asynchronously
- Integration tests are done using docker test to spin up postgres/nats and run tests against them
- Unit tests are done using testify.Suite for business logic related stuff 


## This that can be improved 
- Emit more granular domain events
- This sort of services yields itself to sharding quite well, so that could be a good improvement, sharding could be done by `wallet_id`
- A batch transaction endpoints could be useful depending on usage patterns/backfills 
- Transaction writes can be buffered/queued to create a natural backpressure/rate limiting and to improve write throughput
- Setup resilience patterns (cb, retries, fallbacks and etc.) to make the service more resilient and degradation more gracefully 
- e2e/contract testing 
- Instrument the service with Tracing/profiling/metrics
- Introduce an outbox(er) for event publishing, currently we can't guarantee that the published event actually happened, since the transaction is not guaranteed to be complete after the event is published
- Caching could be introduced to improve read performance (benchmark based on the expected read patters as cache invalidation can be tricky and caching might just complicate things for no real gain)


## Packages in the pkg directory
All packages in this dir are copied over and stripped from things that are not directly used in this project. Here is the list of packages and their purpose:
- config: Configuration package that reads from env vars and provides a typed config struct
- database: Includes a implementation of "transaction manager"/"unit of work" pattern for database transactions
- handler: A generic http handler that makes handlers have stricter input/output types
- pubsub: A generic pubsub package that can be used to publish/subscribe to events, it also includes a nats implementation and a "router" to make handling events easier
- render: A generic render package that can render json responses and handle errors in a consistent way
- requestdecoder: A wrapper around github.com/ggicci/httpin which provides a way to decode various http request params into a struct field, it also provides some json tag validation to prevent common mistakes (not including a json tag and potential field value overwrite exploits)
- server: A small wrapper around http.Server that provides a way to start/stop the server and handle shutdown gracefully 
- sloglog: slog context propagation & helpers
- syncx: provides a generic sync.Map wrapper
- testing: Provides utils useful for testing