-- +goose Up
CREATE TYPE pulse.check_status AS ENUM ('success', 'failure');

CREATE TYPE pulse.check_type AS ENUM (
    'http',
    'tcp',
    'grpc',
    'tls',
    'dns'
    );

CREATE TYPE pulse.check_error_kind AS ENUM (
    'timeout',
    'network',
    'protocol',
    'constraint',
    'internal',
    'unknown'
    );

-- +goose Down
DROP TYPE IF EXISTS pulse.check_error_kind;
DROP TYPE IF EXISTS pulse.check_type;
DROP TYPE IF EXISTS pulse.check_status;
