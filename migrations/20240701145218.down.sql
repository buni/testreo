-- reverse: modify "wallets" table
ALTER TABLE "public"."wallets" ALTER COLUMN "reference_id" TYPE uuid;
