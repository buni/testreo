-- create "wallet_events" table
CREATE TABLE "public"."wallet_events" (
  "id" uuid NOT NULL,
  "version" bigint NOT NULL,
  "transfer_id" text NOT NULL,
  "reference_id" text NOT NULL,
  "wallet_id" uuid NOT NULL,
  "amount" numeric NOT NULL DEFAULT 0,
  "event_type" text NOT NULL,
  "transfer_status" text NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT statement_timestamp(),
  PRIMARY KEY ("id"),
  CONSTRAINT "wallet_events_amount_check" CHECK (amount >= (0)::numeric)
);
-- create index "idx_wallet_events_wallet_id_transfer_id_event_type" to table: "wallet_events"
CREATE UNIQUE INDEX "idx_wallet_events_wallet_id_transfer_id_event_type" ON "public"."wallet_events" ("wallet_id", "transfer_id", "event_type");
-- create "wallet_projections" table
CREATE TABLE "public"."wallet_projections" (
  "wallet_id" uuid NOT NULL,
  "balance" numeric NOT NULL,
  "pending_debit" numeric NOT NULL,
  "pending_credit" numeric NOT NULL,
  "last_event_id" uuid NOT NULL,
  "created_at" timestamp NULL DEFAULT statement_timestamp(),
  "updated_at" timestamp NULL DEFAULT statement_timestamp(),
  PRIMARY KEY ("wallet_id")
);
-- create "wallets" table
CREATE TABLE "public"."wallets" (
  "id" uuid NOT NULL,
  "reference_id" text NOT NULL,
  "created_at" timestamp NULL DEFAULT statement_timestamp(),
  "updated_at" timestamp NULL DEFAULT statement_timestamp(),
  PRIMARY KEY ("id"),
  CONSTRAINT "wallets_reference_id_key" UNIQUE ("reference_id")
);
-- create index "idx_wallet_user_id" to table: "wallets"
CREATE INDEX "idx_wallet_user_id" ON "public"."wallets" ("reference_id");
