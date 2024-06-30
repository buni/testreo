# Project structure


# Decisions 
- Domain events are not emitted by the service, because the service itself works in a syncronus way, I've decided to skip the event buss enterily. This includes domain events, which can be useful for other services to act on (things such as rewards, referal bonuses and etc.)
- Wallet projects are done/updated synchronously during mutations on the transactions api endpoints, this can be done async but I've decided to go with the simplest approach, which should still be performant enough.
- Projections are stored in a separate table instead of a separate db/datastore, again this is for simplicty sake and could easily be changed
- Completing/reverting a transaction will affect the transaction even if it got to a terminal state with a previous event/mutation, there was no explicit requirement for this behaviour so I've decided to go with the simplest approach.
- The service by default doesn't require an idempotency key (a randomly generated one is used), but it can be provided with the caveat that it has to be unique in the context of all transactions in the given wallet
- UUIDv7 is used across the board as it provides a k-sortable time based UUID, which is useful for sorting/ordering
- Atlas is used for database migrations so the migrations/development process has a few extra steps 
- I've copied over a couple of packages I typically use in my personal projects, they typically live in a monorepo, but for simpilicy sake they are copied over here (excluding tests)
- Dockertest is used for "integration"/db tests 
- By default transactions are created in a pending state, and this requires explicit complete/revert calls to change the state, this can be overriden by passing setting the status to complete when creating the transaction
- Pending debit/credit in projections 

# This that can be improved 
- Emit domain events
- Make projections build/rebuild async (I would benchmark both approaches as depending on the size/expected size there might not be a whole lot of difference)
- This sort of services yeilds itself to sharding quite well, so that could be a good improvement, shardings could be done by `wallet_id`
- A batch transaction endpoints could be useful depending on use patterns/backfills 
- Transaction writes can be buffered/queued to create a natural backpressure/rate limiting and to improve write throughput
- Setup resilience patterns (cb, retries, fallbacks and etc.) to make service the service more resilient and degradation more gracefully 
- e2e/contract testing 
- Instrument the service with Tracing/profiling/metrics
- Introduce an outbox(er) along domain events
- Caching could be introduced to improve read performance (benchmark based on the expected read patters as cache invalidation can be tricky and caching might just complicate things for no real gain)
-  