-- +goose Up
CREATE TABLE IF NOT EXISTS pulse.check_executions
(
    id             BIGSERIAL PRIMARY KEY,
    execution_id   uuid                                             NOT NULL,
    check_id       TEXT                                             NOT NULL,
    service_id     TEXT                                             NOT NULL,
    status         pulse.check_status     DEFAULT 'success'         NOT NULL,
    check_type     pulse.check_type                                 NOT NULL,
    started_at     TIMESTAMPTZ                                      NOT NULL,
    finished_at    TIMESTAMPTZ                                      NOT NULL,
    duration       BIGINT                                           NOT NULL,
    attempts_total INTEGER                DEFAULT 1                 NOT NULL,
    error_kind     pulse.check_error_kind DEFAULT NULL,
    error_message  TEXT                   DEFAULT NULL,
    details        JSONB                  DEFAULT NULL,
    created_at     TIMESTAMPTZ            DEFAULT CURRENT_TIMESTAMP NOT NULL,

    CONSTRAINT finished_check CHECK (finished_at >= started_at),
    CONSTRAINT attempts_check CHECK (attempts_total >= 1)
);

CREATE UNIQUE INDEX IF NOT EXISTS check_executions_execution_id_uidx ON pulse.check_executions (execution_id);
CREATE INDEX IF NOT EXISTS check_executions_check_id_idx ON pulse.check_executions (check_id);
CREATE INDEX IF NOT EXISTS check_executions_service_id_idx ON pulse.check_executions (service_id);

CREATE TRIGGER check_executions_prohibit_update
    BEFORE UPDATE
    ON pulse.check_executions
    FOR EACH ROW
EXECUTE FUNCTION pulse.update_prohibited();

CREATE TRIGGER check_executions_prohibit_delete
    BEFORE DELETE
    ON pulse.check_executions
    FOR EACH ROW
EXECUTE FUNCTION pulse.delete_prohibited();

-- +goose Down
DROP TRIGGER IF EXISTS check_executions_prohibit_delete ON pulse.check_executions;
DROP TRIGGER IF EXISTS check_executions_prohibit_update ON pulse.check_executions;
DROP INDEX IF EXISTS check_executions_service_id_idx;
DROP INDEX IF EXISTS check_executions_check_id_idx;
DROP INDEX IF EXISTS check_executions_execution_id_uidx;
DROP TABLE IF EXISTS pulse.check_executions;
