-- Extend the quote state machine with a 'negociation' state (sent → negociation
-- → validated/drop/sent). The original CHECK was added inline with the column,
-- so Postgres gave it an auto-generated name; drop it by discovering that name,
-- then re-create the constraint including the new value.
DO $$
DECLARE
    constraint_name TEXT;
BEGIN
    SELECT con.conname INTO constraint_name
    FROM pg_constraint con
    JOIN pg_class rel ON rel.oid = con.conrelid
    WHERE rel.relname = 'quotes'
      AND con.contype = 'c'
      AND pg_get_constraintdef(con.oid) LIKE '%state%';

    IF constraint_name IS NOT NULL THEN
        EXECUTE format('ALTER TABLE quotes DROP CONSTRAINT %I', constraint_name);
    END IF;
END
$$;

ALTER TABLE quotes
    ADD CONSTRAINT quotes_state_check
    CHECK (state IN ('draft', 'sent', 'negociation', 'validated', 'drop'));
