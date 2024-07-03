-- reverse: create index "idx_outbox_messages_status_publisher_type" to table: "outbox_messages"
DROP INDEX "public"."idx_outbox_messages_status_publisher_type";
-- reverse: create "outbox_messages" table
DROP TABLE "public"."outbox_messages";
