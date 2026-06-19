-- Backfill client_type for pre-existing clients. Heuristic: a client carrying a
-- SIREN, VAT number or company name is treated as a business; everyone else is
-- an individual. This is the best signal available from the legacy schema.
UPDATE clients
SET client_type = CASE
    WHEN COALESCE(NULLIF(siren, ''), NULLIF(vat, ''), NULLIF(company, '')) IS NOT NULL
        THEN 'business'
    ELSE 'individual'
END
WHERE client_type IS NULL;
