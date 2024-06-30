CREATE TABLE wallets (
    id uuid PRIMARY KEY,
    reference_id uuid NOT NULL UNIQUE,
    -- for simplicity sake lets assume that we can have only one wallet per user/reference_id
    name text NOT NULL,
    currency text NOT NULL,
    created_at timestamp DEFAULT statement_timestamp(),
    updated_at timestamp DEFAULT statement_timestamp()
);

CREATE INDEX idx_wallet_user_id ON wallets (reference_id);

CREATE TABLE wallet_projections (
    wallet_id uuid PRIMARY KEY,
    balance decimal NOT NULL,
    last_sequence bigint NOT NULL,
    last_event_id uuid NOT NULL,
    created_at timestamp DEFAULT statement_timestamp(),
    updated_at timestamp DEFAULT statement_timestamp()
);

CREATE TABLE wallet_events (
    id uuid PRIMARY KEY,
    sequence bigint NOT NULL,
    version bigint NOT NULL,
    transfer_id uuid text NOT NULL,
    reference_id uuid text NOT NULL,
    wallet_id uuid NOT NULL,
    amount decimal NOT NULL DEFAULT 0 CHECK (amount >= 0),
    event_type text NOT NULL,
    transfer_status text NOT NULL,
    created_at timestamp NOT NULL DEFAULT statement_timestamp()
);

CREATE UNIQUE INDEX idx_wallet_events_wallet_id_transfer_id_event_type ON wallet_events (wallet_id, transfer_id, event_type);