-- create "outbox_messages" table
CREATE TABLE "public"."outbox_messages" (
  "id" uuid NOT NULL,
  "payload" jsonb NOT NULL,
  "publisher_type" text NOT NULL,
  "publisher_options" jsonb NOT NULL,
  "status" text NOT NULL,
  "created_at" timestamp NOT NULL,
  "updated_at" timestamp NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "idx_outbox_messages_status_publisher_type" to table: "outbox_messages"
CREATE INDEX "idx_outbox_messages_status_publisher_type" ON "public"."outbox_messages" ("status", "publisher_type");
