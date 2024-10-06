CREATE TABLE IF NOT EXISTS event
(
    id uuid UNIQUE NOT NULL, -- leave for client to set
    sequence bigserial NOT NULL,
    type varchar(255) NOT NULL,
    data jsonb NOT NULL,
    meta jsonb NULL,
    causation_event_id varchar(512) NULL, -- TODO - Check
    correlation_event_id varchar(512) NULL, -- TODO - Check
    stream_id varchar(512) NOT NULL, -- aggregate id
    stream_version int NOT NULL, -- aggregate version
--     stream_name text NOT NULL, -- aggregate name
    occurred_on timestamptz NOT NULL DEFAULT (now() at time zone 'utc'),
    PRIMARY KEY (stream_id, stream_version)
);