-- Mark the 27 EU member states (France included; the FR exclusion for OSS lives
-- in the invoice resolution logic, keyed on code='FR', not in this flag).
UPDATE countries
SET is_eu = TRUE
WHERE code IN (
    'AT','BE','BG','HR','CY','CZ','DK','EE','FI','FR','DE','GR','HU','IE',
    'IT','LV','LT','LU','MT','NL','PL','PT','RO','SK','SI','ES','SE'
);
