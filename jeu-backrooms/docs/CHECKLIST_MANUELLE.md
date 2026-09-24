# Checklist de vérification manuelle (characterization tests)

Ce projet n'a **pas de tests automatisés** : cette checklist les
remplace. Elle capture le comportement **actuel** du jeu, textes
compris, mot pour mot. Elle sert de référence avant/après chaque
itération de refactoring : si une seule chaîne ou une seule valeur
diffère après un refacto, c'est une régression.

**Mode d'emploi**
1. `cd src && go run .`
2. Dérouler les scénarios dans l'ordre, cocher au fur et à mesure.
3. Noter dans la colonne « Observé » ce qui s'affiche réellement
   (surtout les nombres : ils dépendent de la classe choisie).
4. Après un refacto, rejouer **au minimum** les scénarios listés dans
   la section « Scénarios à rejouer » de l'itération concernée.

Les textes en `police à chasse fixe` doivent apparaître **exactement**
ainsi (ponctuation et espaces compris).

---

## S1 — Menu et création du personnage

| Étape | Attendu | Observé |
|---|---|---|
| Lancer le jeu | Menu : `Jouer`, `Quitter` (pas de `Continuer` à la 1re partie) | |
| Saisir `jean-luc42` | Le champ affiche `Jeanluc` (chiffres et tiret ignorés, majuscule auto) | |
| Entrée | Écran des 3 classes : `Survivant`, `Ancien Résident`, `Chasseur de niveaux` | |
| Choisir Survivant, Entrée | En jeu. HUD : `Niv. 1`, PV `50/100`, EN `100/100`, `30 pièces`, `Équip. 0/3` | |
| Échap puis retour au menu | Le menu affiche maintenant `Continuer` en premier | |

> PV de départ = **moitié** des PV max : 50/100 (Survivant), 40/80
> (Ancien Résident), 60/120 (Chasseur de niveaux).

## S2 — Déplacement, collisions, transitions

| Étape | Attendu | Observé |
|---|---|---|
| Aller contre un mur | Le joueur est bloqué, ne traverse pas | |
| Longer un mur en diagonale | Il **glisse** le long du mur (pas de blocage sec) | |
| Sortir par une porte | Changement de salle, entrée à la **même hauteur** que la sortie | |
| F1 | Affiche `scene: level1_rXcY` et la liste des sorties | |

## S3 — Inventaire

| Étape | Attendu | Observé |
|---|---|---|
| `E` | Panneau `INVENTAIRE`, compteur `3 / 10`, 3x Almond Water | |
| Sélectionner un emplacement vide | `Emplacement vide` | |
| Sélectionner Almond Water | Description : `Eau légèrement sucrée. Rend 50 PV.` | |
| Entrée sur Almond Water (PV 50/100) | Toast `Almond Water bue : +50 PV`, HUD passe à `100/100`, compteur `2 / 10` | |
| Entrée à PV pleins | Toast `Tes PV sont déjà au maximum.` et **la bouteille n'est pas consommée** | |
| F4 | Toast `Inventaire agrandi : 20 places`, grille sur 10 colonnes | |
| F4 trois fois de plus | 3e OK (`40 places`), la 4e affiche `Amélioration maximale atteinte` | |
| Bas de panneau | `Flèches : choisir   ·   Entrée : utiliser   ·   E : fermer` | |

## S4 — Marchand (salle de départ)

| Étape | Attendu | Observé |
|---|---|---|
| S'approcher du PNJ | Bulle `R : parler au marchand` | |
| `R` | Panneau `LE MARCHAND`, onglets `ACHETER` / `VENDRE`, compteur de pièces | |
| Liste d'achat | `Almond Water` 10 po, `Potion de poison` 20 po, `Boule de feu` 120 po, `Souvenir d'Almond Water` 60 po | |
| Acheter Almond Water | Toast `Acheté : Almond Water`, pièces −10 | |
| Acheter la Boule de feu à 30 po | Toast `Pas assez de pièces.` | |
| Acheter la Boule de feu avec assez d'or | Toast `Acheté : Boule de feu`, la ligne devient `déjà appris` | |
| Racheter le grimoire | Toast `Ce sort est déjà appris.` | |
| Acheter avec inventaire plein | Toast `Inventaire plein.` | |
| `Tab` → `VENDRE` sans loot | `(rien à vendre pour le moment)` | |
| Vendre un Tuyau rouillé | Toast `Vendu pour 16 pièces` | |
| Échap | Retour au jeu, le joueur n'a pas bougé | |

## S5 — Monstres et déclenchement du combat

| Étape | Attendu | Observé |
|---|---|---|
| Explorer plusieurs salles | Des monstres noirs filamenteux, **plus grands** que le joueur | |
| S'approcher d'un monstre | Un `!` apparaît au-dessus de lui, il se met à poursuivre | |
| Fuir en ligne droite | Le joueur **distance** le monstre (vitesse 0,7x la sienne) | |
| S'éloigner beaucoup | Le `!` disparaît, le monstre revient en errance | |
| Se laisser toucher | Fondu au noir puis écran de combat | |
| Fuir le combat, revenir | Pas de re-déclenchement immédiat (délai de grâce ~1,5 s) | |
| Changer de salle | Même délai de grâce à l'arrivée | |
| Aucune salle | Jamais de monstre dans la salle de **départ** ni celle du **boss** | |

## S6 — Combat : tours, attaques, énergie

| Étape | Attendu | Observé |
|---|---|---|
| Entrée en combat | Monstre en haut à droite, joueur **de dos** en bas à gauche | |
| Boîte de texte | `Que fait <Nom> ?` puis `Attaque` / `Sac` / `Fuite` | |
| Menu Attaque | `Coup de poing`, `Griffe électrique (15 EN)`, puis `Boule de feu (verrouillé)` et `Souvenir d'Almond Water (verrouillé)` tant que leurs grimoires ne sont pas achetés | |
| Choisir la Boule de feu verrouillée | Toast `Ce sort n'est pas encore appris.` | |
| Griffe électrique sans énergie | Toast `Pas assez d'énergie.` | |
| Souvenir d'Almond Water (appris) à PV pleins | Toast `Tes PV sont déjà au maximum.`, le tour n'est pas perdu | |
| Souvenir d'Almond Water (appris), PV entamés | Message `<Nom> utilise Souvenir d'Almond Water : +X PV.` (30 max) | |
| Coup de poing | Message `<Nom> utilise Coup de poing : -X PV.` + flash blanc | |
| Riposte du monstre | `Le monstre attaque : -X PV.` + secousse d'écran | |
| Au 3e tour | Le message contient `(motif renforcé)` et les dégâts sont **doublés** | |
| Entre deux tours | L'énergie remonte de **+12** | |
| Messages | Ils défilent **un par un** avec Entrée (`Entrée : continuer`) | |
| Échap dans un sous-menu | Retour au menu principal du combat | |

## S7 — Potion de poison

| Étape | Attendu | Observé |
|---|---|---|
| Sac → `Potion de poison` | `Potion de poison jetée sur le monstre !` | |
| Tours suivants | `Le poison ronge le monstre (-8 PV).` pendant **4 tours** au total | |
| Si le poison achève le monstre | `Le monstre succombe au poison !` | |
| Utiliser la potion hors combat | `Se garde pour empoisonner un monstre en combat (Sac).` | |

## S8 — Victoire, loot, ramassage

| Étape | Attendu | Observé |
|---|---|---|
| Tuer un monstre | `Le monstre est vaincu !` puis retour au jeu | |
| Sol | 0 à **3 types différents** d'objets maximum, jamais plus | |
| Marcher dessus | Toast `Ramassé : <objet>`, l'objet entre dans l'inventaire | |
| Marcher dessus, inventaire plein | Toast `Inventaire plein !` et **l'objet reste au sol** | |
| Attendre ~8 s | Un nouveau monstre réapparaît ailleurs (jamais dans la salle du joueur) | |

## S9 — Forgeron

| Étape | Attendu | Observé |
|---|---|---|
| Trouver le PNJ (salle aléatoire) | Bulle `R : parler au forgeron` | |
| `R` | Panneau `LE FORGERON`, 3 recettes | |
| Recette sans matériaux | Les quantités manquantes s'affichent en **rouge** (`0/2`) | |
| Fabriquer sans matériaux | Toast `Matériaux manquants.` | |
| Fabriquer sans or | Toast `Pas assez de pièces.` | |
| Fabriquer le Casque en tuyau (2x Tuyau rouillé + 30 po) | Toast `Casque en tuyau fabriqué et équipé !`, PV max **+20**, HUD `Équip. 1/3`, matériaux et or déduits | |
| Refabriquer la même pièce | L'ancienne revient **dans l'inventaire** (échange) | |
| Utiliser la pièce depuis l'inventaire | `Casque en tuyau équipé(e) (Tête).` | |

## S10 — Mort et résurrection

| Étape | Attendu | Observé |
|---|---|---|
| Perdre un combat | `Tu perds connaissance...` | |
| Entrée | Retour **à la salle de départ**, toast `Tu es mort... Réveil à moitié de tes PV.` | |
| HUD | PV = **moitié** des PV max, énergie pleine | |

## S11 — Boss (salle du haut, au milieu)

| Étape | Attendu | Observé |
|---|---|---|
| Entrer dans `level1_r0c2` | Une silhouette **immobile**, plus grande que les monstres | |
| La toucher | Dialogue `Une présence immense bloque le passage.` / `Affronter le boss des Backrooms ?` avec **`Non` présélectionné** | |
| Choisir `Non` | Retour au jeu, pas de combat imposé | |
| Choisir `Oui` → menu Fuite | `Impossible de fuir ce combat !` et le combat continue | |
| Combattre **sans équipement** | Chaque coup du boss inflige `+ aura` ; impossible de tenir même en buvant une potion par tour | |
| Combattre avec **3 pièces d'équipement** | Plus de mention `aura` dans les messages, combat gagnable | |
| Victoire | Écran `TU AS ÉCHAPPÉ AUX BACKROOMS` | |

## S12 — Expérience et niveau

| Étape | Attendu | Observé |
|---|---|---|
| Cumuler 100 XP (5 monstres) | Toast `Niveau supérieur !` | |
| HUD | `Niv. 2`, PV max **+10**, énergie max **+5** | |

---

## Playthrough complet (critère de fin de refacto)

Création perso → achat chez le marchand → combat et loot → forge d'une
pièce → équipement → montée de niveau → boss → écran de victoire.
Tout doit s'afficher **à l'identique** avant et après refacto.

## Scénarios à rejouer selon la zone touchée

| Zone refactorée | Scénarios minimum |
|---|---|
| Personnage / HUD | S1, S3, S12 |
| Entrées clavier / menus | S1, S3, S4, S6, S9, S11 |
| Inventaire | S3, S8, S9 |
| Marchand / Forgeron | S4, S9 |
| Dessin (UI) | S1, S3, S4, S6, S9, S11 |
| Boucle de jeu / monde | S2, S5, S8, S10 |
| Combat | S6, S7, S8, S10, S11 |
