-- Free: pas de B2B, pas de clients liés, pas de catalogue de frais, pas de projets
UPDATE plans
SET features = features
    || '{"b2b_invoicing":0,"max_linked_clients":0,"fees_catalog":0,"max_projects":0}'::jsonb
WHERE tier = 'free';

-- Pro: B2B + 5 clients liés + catalogue de frais + 10 projets
UPDATE plans
SET features = features
    || '{"b2b_invoicing":1,"max_linked_clients":5,"fees_catalog":1,"max_projects":10}'::jsonb
WHERE tier = 'pro';

-- Enterprise: tout illimité
UPDATE plans
SET features = features
    || '{"b2b_invoicing":1,"max_linked_clients":-1,"fees_catalog":1,"max_projects":-1}'::jsonb
WHERE tier = 'enterprise';
