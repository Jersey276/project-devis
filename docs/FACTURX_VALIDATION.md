# Validation Factur-X

La génération Factur-X (export d'une facture émise avec `?facturx=1`) produit un
PDF/A-3b portant le XML CII EN 16931 en pièce jointe `factur-x.xml`
(`/AFRelationship Alternative`, entrée catalogue `/AF`, XMP `urn:factur-x` +
extension schema PDF/A).

Deux niveaux de validation, **hors pipeline Go** (outils externes) :

## 1. Conformité PDF/A-3b — veraPDF

Génère un PDF Factur-X de test puis valide-le. Depuis `backend/export`, avec la
stack `docker compose` up (Gotenberg sur le réseau `backend_default`) :

```bash
# 1. Générer un échantillon assemblé dans testdata/ (test utilitaire gardé)
#    Réactiver temporairement un petit test qui écrit le PDF, ou exporter une
#    vraie facture via l'API (?facturx=1) et récupérer le fichier.

# 2. Valider la conformité PDF/A-3b
docker run --rm -v "$PWD/services/facturxpdf/testdata":/data \
  verapdf/cli:latest --flavour 3b /data/facturx_sample.pdf
```

Attendu : `isCompliant="true"`, `failedRules="0"`.

Les deux exigences PDF/A spécifiques à Factur-X, déjà gérées par
`services/facturxpdf` :

- **Extension schema XMP** (ISO 19005-3 §6.6.2.3) : le namespace `urn:factur-x`
  est déclaré via un bloc `pdfaExtension:schemas` (sinon les propriétés `fx:*`
  sont rejetées).
- **MIME du fichier embarqué** (§6.8) : le stream EmbeddedFile porte
  `/Subtype text/xml` (pdfcpu ne le pose pas par défaut).

## 2. Conformité Factur-X EN 16931 — Mustangproject

Valide le couple PDF + XML CII contre le profil EN 16931 :

```bash
docker run --rm -v "$PWD/services/facturxpdf/testdata":/data \
  ghcr.io/zugferd/mustangproject:latest --action validate \
  --source /data/facturx_sample.pdf
```

(À défaut d'image, utiliser le JAR Mustang `--action validate`.) Attendu : XML
valide contre le schéma et les règles métier EN 16931.

## Intégration Go (chaîne complète, sans validation externe)

Le test `services/facturxpdf/integration_test.go` (gardé par `FACTURX_INTEGRATION`)
exécute HTML → Gotenberg PDF/A-3b → `Assemble`, et vérifie l'attachment, le
maintien du marqueur PDF/A-3 et la présence du XMP `urn:factur-x` :

```bash
docker run --rm --network backend_default \
  -v "$PWD":/app -w /app \
  -e FACTURX_INTEGRATION=1 -e GOTENBERG_ADDRESS=http://gotenberg:3000 \
  golang:1.25.6 go test ./services/facturxpdf/ -run EndToEnd -v
```
