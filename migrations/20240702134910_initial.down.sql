-- reverse: create index "idx_wallet_user_id" to table: "wallets"
DROP INDEX "public"."idx_wallet_user_id";
-- reverse: create "wallets" table
DROP TABLE "public"."wallets";
-- reverse: create "wallet_projections" table
DROP TABLE "public"."wallet_projections";
-- reverse: create index "idx_wallet_events_wallet_id_transfer_id_event_type" to table: "wallet_events"
DROP INDEX "public"."idx_wallet_events_wallet_id_transfer_id_event_type";
-- reverse: create "wallet_events" table
DROP TABLE "public"."wallet_events";
