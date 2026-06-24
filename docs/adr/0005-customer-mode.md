# ADR 0005 - Mode customer (vue client)

- Statut: Accepte
- Date: 2026-06-24

## Contexte

L'application est utilisee par deux profils distincts :

- **Prestataire** (`provider`) : cree les devis, gere les clients, les factures,
  les echeanciers et les templates.
- **Client** (`customer`) : consulte les devis qui lui sont destines, en lecture
  seule. Il n'a pas acces aux outils de gestion.

Sans un mecanisme de bascule explicite, un prestataire qui partage son espace
avec un client doit gerer les droits au niveau des routes, ce qui devient lourd.
Le besoin est de proposer une "vue allégée" accessible depuis la meme session,
sans changement d'authentification.

## Decision

1. **Un contexte React `ModeProvider`** expose `mode: "provider" | "customer"`,
   `isProvider`, `isCustomer` et `setMode`. Il est monte dans le layout
   authentifié (`app/(app)/layout.tsx`).

2. **Persistance par cookie** : le mode est lu au montage depuis le cookie
   `user-mode` (max-age 1 an, `SameSite=Lax`, path `/`). Chaque appel a
   `setMode` reecrit le cookie. Le cookie est absent par defaut → mode
   `provider`.

3. **Toggle dans la sidebar** : un bouton `data-slot="mode-toggle"` dans le
   footer de `AppSidebar` bascule entre les deux modes. Il porte
   `data-active="true"` quand `isCustomer` est vrai. Les libelles sont
   localises dans `fr.json > nav.modeToggle`.

4. **Items de navigation** : les items ayant `modes: ["provider"]` sont masques
   en mode customer. Le seul item visible en mode customer est "Devis" (pas de
   contrainte `modes`). Aucun item `modes: ["customer"]` n'existe pour l'instant.

5. **Comportement par feature** :

   | Feature | Mode provider | Mode customer |
   |---|---|---|
   | Liste des devis | fetch + filtres + pagination | vide (fetch skipé), empty state |
   | Detail d'un devis | lecture/ecriture | lecture seule (inputs disabled) |
   | /quote/create | accessible | redirect vers /quote |
   | Bouton "Nouveau devis" | visible | masqué |
   | Actions (PDF, echéancier, template) | visibles | masquées |
   | Sidebar : Factures, Clients, etc. | visibles | masquées |

6. **Cypress** : `cy.visitAs(mode, url)` set le cookie `user-mode` avant
   `cy.visit`. La suite `front/cypress/e2e/quote-customer.cy.ts` couvre tous
   les cas ci-dessus.

## Contraintes de l'implementation actuelle

- Le mode customer ne filtre pas les devis par "client associe" : la liste est
  vide. Cette limitation est assumee en attendant un modele de liaison
  utilisateur ↔ client cote backend.
- Il n'existe pas de route `/customer/*` separee : c'est le meme frontend avec
  une vue filtree.
- Le cookie n'est pas signe : un utilisateur peut se mettre en mode customer
  lui-meme. Ce n'est pas un probleme de securite (le mode ne donne pas acces a
  plus de donnees, il en cache).

## Consequences positives

- Aucun changement d'authentification requis pour demonstrer l'interface cliente.
- Persistance transparente entre navigations grace au cookie.
- Les composants existants pilotent leur comportement via `isCustomer` sans
  logique de route separee.
- `visitAs` dans Cypress permet de tester le mode proprement sans simuler un
  toggle UI.

## Consequences negatives

- La liste des devis en mode customer est toujours vide : pas utile en production
  tant que le lien utilisateur ↔ client n'est pas implemente.
- Le mode est global a la session : si deux onglets sont ouverts, changer le mode
  dans l'un affecte le cookie de l'autre (lecture au prochain montage seulement).

## Alternatives ecartées

1. **Routes separees `/customer/*`** : duplique le code et complique la
   navigation ; ecarté au profit d'un contexte React.
2. **Parametre URL `?mode=customer`** : visible et manipulable dans l'URL, pas
   persistant entre pages sans logique supplementaire.
3. **Lecture du role depuis le JWT** : le role est lie au compte, pas a la
   session en cours. Un prestataire ne peut pas "devenir" son propre client via
   son JWT.

## References

- `front/lib/mode-context.tsx`
- `front/components/custom/app-sidebar.tsx`
- `front/components/quote/quote-list-table.tsx`
- `front/components/quote/quote-form.tsx`
- `front/components/quote/customer-redirect.tsx`
- `front/components/quote/new-quote-button.tsx`
- `front/cypress/e2e/quote-customer.cy.ts`
- `front/cypress/support/commands.ts`
