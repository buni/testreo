-- reverse: modify "wallets" table
ALTER TABLE "public"."wallets" ADD COLUMN "name" text NOT NULL;
