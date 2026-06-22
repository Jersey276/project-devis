# ADR 0004 - Pattern de filtres dans les tableaux

- Statut: Accepte
- Date: 2026-06-22

## Contexte

Plusieurs tableaux dans l'application (logs, factures, devis, etc.) exposent des
filtres. Sans convention etablie, chaque tableau risquait d'avoir sa propre
disposition : certains filtres en ligne au-dessus du tableau, d'autres dans des
menus deroulants, d'autres encore dans des barres d'outils. Cela nuit a la
coherence de l'interface et complique la reutilisation des composants.

La premiere implementation concrete est le tableau des logs d'audit, qui combine
une recherche textuelle (URL, identifiant utilisateur) et des filtres plus
structures (statuts HTTP, plage de dates).

## Decision

1. **Barre de recherche toujours en ligne**, a gauche au-dessus du tableau. Elle
   couvre tous les champs texte libres que le tableau expose (URL, ID utilisateur,
   nom, reference, etc.). Un seul champ, une seule valeur envoyee ; c'est le
   backend qui decide comment distribuer la valeur aux colonnes concernees (OR
   logique).

2. **Tous les autres filtres dans une sidebar a droite** (`FilterSidebar`,
   `front/components/ui/filter-sidebar.tsx`). Le bouton d'ouverture de la sidebar
   est place a cote de la barre de recherche. Il affiche un badge numerique
   indiquant le nombre de filtres actifs (hors recherche). La sidebar contient
   autant de sections que de categories de filtres (`FilterSidebarSection`).

3. **`FilterSidebar` est generique** : il ne connait pas le domaine metier. Il
   recoit ses libelles (`triggerLabel`, `title`, `resetLabel`) en props depuis le
   composant appelant, qui les source de son propre namespace i18n.

4. **La recherche est separee des filtres sidebar** dans le type de donnees
   (`LogFilters.search` vs les autres champs). Le parent du composant de filtres
   reste la source de verite ; le composant est entierement controle (pas d'etat
   interne draft).

## Consequences positives

- Interface coherente : un utilisateur qui apprend le pattern sur les logs le
  retrouve sur les factures, les devis, etc.
- `FilterSidebar` reutilisable sans couplage i18n ou metier.
- La barre de recherche reste accessible sans ouvrir la sidebar (cas d'usage le
  plus frequent).
- Le badge actif sur le bouton sidebar signale clairement qu'un filtre est en
  place meme quand la sidebar est fermee.

## Consequences negatives

- Les filtres sidebar ne sont pas visibles d'emblee : un filtre de date actif
  peut passer inapercus si l'utilisateur ne remarque pas le badge.
- La separation recherche / sidebar implique deux zones d'interaction ; pour
  des tableaux tres simples (un seul filtre texte), la sidebar est du sur-
  engineering.

## Alternatives ecartees

1. Tous les filtres en ligne au-dessus du tableau : encombre l'interface, ne passe
   pas a l'echelle quand le nombre de filtres augmente.
2. Filtres dans un panneau collapse au-dessus du tableau : pattern moins standard
   sur les applications de gestion, et ne prepare pas bien la reutilisation via
   un composant commun.
3. Etat interne draft avec bouton "Appliquer" : ajoute une etape a chaque
   changement de filtre. Abandonne au profit d'une mise a jour immediate (chaque
   interaction appelle `onChange` directement).

## References

- `front/components/ui/filter-sidebar.tsx`
- `front/components/admin/logs/logs-filters.tsx` (premiere implementation)
- ADR 0001 (architecture generale)
