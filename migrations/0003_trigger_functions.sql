-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION pulse.update_prohibited() RETURNS trigger AS
$$
BEGIN
    RAISE EXCEPTION 'Update on table %.% is prohibited', TG_TABLE_SCHEMA, TG_TABLE_NAME
        USING ERRCODE = 'P1001';
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION pulse.delete_prohibited() RETURNS trigger AS
$$
BEGIN
    RAISE EXCEPTION 'Delete on table %.% is prohibited', TG_TABLE_SCHEMA, TG_TABLE_NAME
        USING ERRCODE = 'P1002';
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
DROP FUNCTION pulse.delete_prohibited();
DROP FUNCTION pulse.update_prohibited();
