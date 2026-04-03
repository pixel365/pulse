-- +goose Up
CREATE TABLE IF NOT EXISTS pulse.check_states
(
    id                    BIGSERIAL PRIMARY KEY,
    check_id              TEXT                                               NOT NULL,
    service_id            TEXT                                               NOT NULL,
    check_type            pulse.check_type         DEFAULT 'http'            NOT NULL,
    status                pulse.check_state_status DEFAULT 'unknown'         NOT NULL,
    last_execution_id     uuid                                               NOT NULL,
    last_status           pulse.check_status       DEFAULT 'success'         NOT NULL,
    last_error_kind       pulse.check_error_kind   DEFAULT NULL,
    last_error_message    TEXT                     DEFAULT NULL,
    last_duration         BIGINT                                             NOT NULL,
    last_details          JSONB                    DEFAULT NULL,
    last_success_at       TIMESTAMPTZ              DEFAULT NULL,
    last_failure_at       TIMESTAMPTZ              DEFAULT NULL,
    consecutive_successes INT                      DEFAULT 0                 NOT NULL,
    consecutive_failures  INT                      DEFAULT 0                 NOT NULL,
    updated_at            TIMESTAMPTZ              DEFAULT CURRENT_TIMESTAMP NOT NULL,

    CONSTRAINT check_states_check_id_service_id_unique UNIQUE (check_id, service_id),
    CONSTRAINT check_consecutive_failures_positive CHECK (consecutive_failures >= 0),
    CONSTRAINT check_consecutive_successes_positive CHECK (consecutive_successes >= 0)
);

CREATE INDEX IF NOT EXISTS check_states_check_id_idx ON pulse.check_states (check_id);
CREATE INDEX IF NOT EXISTS check_states_service_id_idx ON pulse.check_states (service_id);

CREATE TRIGGER check_states_prohibit_delete
    BEFORE DELETE
    ON pulse.check_states
    FOR EACH ROW
EXECUTE FUNCTION pulse.delete_prohibited();

-- +goose Down
DROP TRIGGER IF EXISTS check_states_prohibit_delete ON pulse.check_states;
DROP INDEX IF EXISTS check_states_service_id_idx;
DROP INDEX IF EXISTS check_states_check_id_idx;
DROP TABLE IF EXISTS pulse.check_states;
