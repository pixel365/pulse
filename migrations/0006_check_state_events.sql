-- +goose Up
CREATE TABLE pulse.check_state_events
(
    id                    BIGSERIAL PRIMARY KEY,
    execution_id          uuid                                             NOT NULL,
    check_id              TEXT                                             NOT NULL,
    service_id            TEXT                                             NOT NULL,
    check_type            pulse.check_type                                 NOT NULL,
    status                pulse.check_state_status                         NOT NULL,
    last_status           pulse.check_status                               NOT NULL,
    last_error_kind       pulse.check_error_kind DEFAULT NULL,
    last_error_message    TEXT                   DEFAULT NULL,
    last_duration         BIGINT                                           NOT NULL,
    last_details          JSONB                  DEFAULT NULL,
    last_success_at       TIMESTAMPTZ            DEFAULT NULL,
    last_failure_at       TIMESTAMPTZ            DEFAULT NULL,
    consecutive_successes INT                                              NOT NULL,
    consecutive_failures  INT                                              NOT NULL,
    observed_at           TIMESTAMPTZ                                      NOT NULL,
    created_at            TIMESTAMPTZ            DEFAULT CURRENT_TIMESTAMP NOT NULL,

    CONSTRAINT check_state_events_execution_id_uidx UNIQUE (execution_id),
    CONSTRAINT check_state_events_consecutive_failures_positive CHECK (consecutive_failures >= 0),
    CONSTRAINT check_state_events_consecutive_successes_positive CHECK (consecutive_successes >= 0)
);

CREATE INDEX check_state_events_service_check_observed_id_idx
    ON pulse.check_state_events (service_id, check_id, observed_at, id) INCLUDE (status, last_status);

CREATE TRIGGER check_state_events_prohibit_update
    BEFORE UPDATE
    ON pulse.check_state_events
    FOR EACH ROW
EXECUTE FUNCTION pulse.update_prohibited();

CREATE TRIGGER check_state_events_prohibit_delete
    BEFORE DELETE
    ON pulse.check_state_events
    FOR EACH ROW
EXECUTE FUNCTION pulse.delete_prohibited();

-- +goose Down
DROP TRIGGER IF EXISTS check_state_events_prohibit_delete ON pulse.check_state_events;
DROP TRIGGER IF EXISTS check_state_events_prohibit_update ON pulse.check_state_events;
DROP INDEX IF EXISTS check_state_events_service_check_observed_id_idx;
DROP TABLE IF EXISTS pulse.check_state_events;
