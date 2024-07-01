-- reverse: modify "wallets" table
ALTER TABLE "public"."wallets" ADD COLUMN "currency" text NOT NULL;
