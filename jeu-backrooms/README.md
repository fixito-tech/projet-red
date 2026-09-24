# Projet RED — The Backrooms (moteur graphique)

Moteur 2D en Go avec Ebiten : menu, personnage, inventaire, monstres,
combat au tour par tour façon Pokémon, économie (marchand/forgeron),
boss final. Voir [docs/TABLEAU_TACHES.md](docs/TABLEAU_TACHES.md) pour
la correspondance tâche par tâche avec l'énoncé.

## Lancer

Le code, les images et les scripts sont dans `src/` : on lance le jeu
**depuis ce dossier**, sinon il ne trouve pas ses images.

```bash
cd src
go mod tidy
go run .
```

La validation se fait à la main : voir
[docs/CHECKLIST_MANUELLE.md](docs/CHECKLIST_MANUELLE.md).

## Organisation du code

Tous les fichiers sont dans `src/`, dans le même `package main`. Chaque
écran a un fichier pour le clavier (`*_input.go`) et un pour le dessin
(`*_ui.go`) ; les règles du jeu sont dans des fichiers sans suffixe.
Le [guide de compréhension](docs/GUIDE.md) détaille chaque fichier et le
déroulement d'une partie.

```
src/
├── main.go                      point d'entrée : lance le jeu
├── game.go                      struct Game, Update / Draw, messages temporaires
├── config.go                    réglages chiffrés (taille des cases, vitesse…)
├── input.go                     touches des menus (flèches, Entrée…)
│
├── player.go                    struct Character, création, XP, mort
├── inventory.go                 inventaire : ajouter, retirer, utiliser
├── equipment.go                 struct Equipment (tête, torse, pieds)
├── spells.go                    sorts appris / verrouillés
├── economy.go                   marchand et forgeron (prix, recettes)
├── combat.go                    règles du combat au tour par tour
├── monster.go                   struct Monster, apparition, IA, loot
├── boss.go                      boss et son aura
├── scene.go                     salles : chargement, murs, portes
├── world.go                     le monde qui vit pendant l'exploration
├── errors.go                    messages d'erreur personnalisés
│
├── menu_input.go / menu_ui.go         menu titre, création du personnage
├── play_input.go / play_ui.go         exploration des salles
├── inventory_ui.go                    panneau d'inventaire et icônes
├── economy_input.go / economy_ui.go   marchand et forgeron
├── combat_input.go / combat_ui.go     écran de combat, boss, victoire
├── monster_draw.go                    monstres et loot à l'écran
├── ui.go                              outils de dessin, HUD, toast
├── placeholders.go                    visuels de secours si un PNG manque
│
└── assets/                      images et grilles des salles
```

## Commandes

| Touche | Action |
|---|---|
| Flèches ou ZQSD | Se déplacer / naviguer dans les menus |
| Entrée / Espace | Valider |
| **E** | Ouvrir/fermer l'inventaire |
| **R** | Parler au PNJ à proximité (marchand, forgeron) |
| Échap | Retour / fermer un menu |
| Tab | (chez le marchand) basculer Acheter / Vendre |
| F1 | Infos de débogage (scène, sorties) |
| F2–F4 | Touches de test (voir plus bas) |
| F11 / Alt+Entrée | Plein écran |

## Ajouter une scène

Chaque scène = 2 fichiers dans `src/assets/scenes/` :
- `nom.png` : le décor (800x480)
- `nom.txt` : la grille de collisions 25x15 + les sorties

Caractères de la grille : `#` mur, `.` sol, `+` porte, `@` départ.
En-tête du .txt : `# scene: nom` puis `# exit: est -> autre_scene`.

La carte fournie (`level1_r0c0` à `level1_r4c4`) est une grille 5x5 :
le départ est en `level1_r3c2` (alias `level1_start`), le boss tout en
haut au milieu en `level1_r0c2`. Le moteur s'adapte à une autre carte :
si le départ ou la salle du boss n'existent pas, il se replie
proprement (position par défaut au centre de l'écran, pas de boss).

## Remplacer les sprites

Tous les visuels sont générés par des scripts Python (reproductibles),
**avec un visuel de secours dessiné directement dans le code Go si le
fichier PNG est absent** — le jeu tourne donc même sans Pillow.

Les scripts sont dans `src/` et **se lancent depuis ce dossier** : ils
écrivent dans `assets/`, un chemin relatif à `src/`.

| Script | Fichiers produits | Secours Go si absent |
|---|---|---|
| `python3 make_player.py` | `assets/player.png` (personnage) | `placeholderSheet` (placeholders.go) |
| `python3 make_monsters.py` | `assets/monster.png`, `assets/boss.png`, `assets/merchant.png`, `assets/blacksmith.png` | `placeholderMonsterSheet`, `placeholderNPCSheet` (placeholders.go) |

Un fond de menu optionnel peut être mis dans `src/assets/menu.png`.

> Ces deux scripts nécessitent Python 3 + Pillow (`pip install
> pillow`) ; ils n'ont pas pu être exécutés dans cet environnement de
> développement (Python indisponible), mais le jeu a été testé et
> fonctionne avec les visuels de secours. Lance-les de ton côté pour
> obtenir les sprites définitifs, puis vérifie-les visuellement.

## Le personnage

`assets/player.png` — combinaison antiradiation jaune, masque à gaz
noir. Feuille de 160x192 px : 4 colonnes (animation) x 4 lignes (bas,
gauche, droite, haut). Le sprite fait 40x48 px alors qu'une case en
fait 32x32 : il déborde vers le haut et sur les côtés, mais son
empreinte au sol reste d'une case (les collisions ne changent pas).

## Affichage

- La fenêtre s'ouvre en 1600x960 (jeu interne en 800x480, x2).
- **F11** ou **Alt+Entrée** : plein écran / fenêtre.
- Pour démarrer directement en plein écran : `startFullscreen = true`
  dans `config.go`.

## Personnage et inventaire

Après « Jouer » : saisie du nom (lettres uniquement, majuscule ajoutée
automatiquement), puis choix de la classe.

| Classe | Base (énoncé) | PV max | Énergie max | Vitesse (initiative) |
|--------|---------------|--------|-------------|----|
| Survivant | Humain | 100 | 100 | 10 |
| Ancien Résident | Elfe | 80 | 120 | 14 |
| Chasseur de niveaux | Nain | 120 | 80 | 7 |

On démarre niveau 1 avec la moitié de ses PV max, 3 Almond Water et
30 pièces. En jeu : **E** ouvre/ferme l'inventaire (10 emplacements ;
+10 par amélioration, 3 fois maximum). En cas de mort, résurrection à
la salle de départ avec 50 % des PV max (pas d'écran de game over).

### Touches de test (pratiques pour la démo à l'oral)
- **F2** : ramasse un objet de test
- **F3** : prend 15 dégâts et perd 10 d'énergie
- **F4** : agrandit l'inventaire de 10 places
- **F1** : infos de débogage

## Monstres — le « monstre spaghetti »

Jusqu'à **9 monstres** vivent sur toute la carte, apparus loin des
portes et de la salle de départ (jamais dans la salle de départ ni
celle du boss). Un monstre erre au repos ; si le joueur entre dans son
rayon de détection, il le poursuit (plus lentement que lui : on peut
toujours fuir) et affiche un **!** au-dessus de sa tête. Le combat se
déclenche au contact ; un court délai de grâce suit chaque combat et
chaque changement de salle pour éviter un re-déclenchement immédiat.
Quand un monstre meurt, un autre réapparaît progressivement après un
délai (tant qu'on reste sous 9).

À sa mort, un monstre lâche entre 0 et 3 types d'objets différents
(jamais plus), à ramasser en marchant dessus.

## Combat au tour par tour (façon Pokémon)

Transition animée en fondu, puis écran dédié : monstre en haut à
droite, joueur de dos en bas à gauche, barres de vie/énergie, boîte de
texte avec menu **Attaque / Sac / Fuite** (fuite impossible contre le
boss). Les coups qui portent déclenchent un flash et une secousse
d'écran.

Quatre attaques :
| Attaque | Coût | Dégâts | Débloquée |
|---|---|---|---|
| Coup de poing | gratuit | 8–14 | dès le départ |
| Griffe électrique | 15 énergie | 14–20 | dès le départ |
| Boule de feu | 25 énergie | 22–30 | après achat du grimoire chez le marchand |
| Souvenir d'Almond Water | 20 énergie | soigne 30 PV (refusé si PV pleins) | après achat du grimoire chez le marchand (60 po) |

L'énergie remonte de 12 à chaque tour de combat, et d'1 point toutes
les ~40 images hors combat. La **potion de poison** (achetée chez le
marchand, utilisable depuis le Sac en combat) inflige des dégâts sur 4
tours au monstre. Les monstres normaux ont un motif d'attaque : leurs
dégâts doublent tous les 3 tours. La vie d'un monstre normal vaut 1x
les PV max de base du joueur, hors bonus d'équipement (équilibrage
assumé : l'énoncé prévoit 2x, voir « Divergences signalées »).

**Mission bonus « initiative »** : chaque classe et chaque monstre a
une vitesse ; le plus rapide agit en premier dans le tour.

## Économie et artisanat

- **Pièces** : monnaie séparée de l'inventaire, affichée dans le HUD.
- **Marchand** (au point de départ) : achète le loot ramassé sur les
  monstres, vend Almond Water, potion de poison et les grimoires
  « Boule de feu » et « Souvenir d'Almond Water ». Touche **R** pour lui parler, **Tab** pour
  basculer Acheter/Vendre.
- **Forgeron** (dans une salle aléatoire de la carte) : fabrique une
  pièce d'équipement par recette (matériaux + pièces), qui s'équipe
  aussitôt et renvoie l'ancienne pièce dans l'inventaire (échange).

| Recette | Matériaux | Prix | Bonus |
|---|---|---|---|
| Casque en tuyau (tête) | 2x Tuyau rouillé | 30 po | +20 PV max |
| Plastron de moquette (torse) | 3x Morceau de moquette | 50 po | +35 PV max |
| Bottes en fer (pieds) | 2x Barre de fer | 25 po | +15 PV max |

Erreurs personnalisées (voir `errors.go`) : pas assez de pièces,
inventaire plein, matériaux manquants, sort déjà connu, objet non
vendable, objet introuvable.

## Le boss

Immobile, tout en haut au milieu de la carte (`level1_r0c2`). Le
toucher ouvre un dialogue de confirmation (« Affronter le boss ? Oui /
Non ») — pas de combat imposé par surprise. La fuite est impossible
une fois le combat engagé.

**Sans équipement, le boss est structurellement imbattable** : une
aura lui fait infliger, en plus de son coup, des dégâts qui dépassent
toujours ce qu'une Almond Water peut soigner (60 de base contre 50
soignés) — chaque pièce d'équipement portée (tête/torse/pieds) réduit
cette aura de 20, jusqu'à l'annuler complètement à trois pièces. Ce
n'est donc pas un simple indicateur « invincible » : c'est un calcul
de dégâts qui rend toute stratégie (potions, attaques) insuffisante
tant qu'on n'a pas au moins un peu d'équipement, et qui devient
gagnable une fois complètement équipé (scénario S11 de la
[checklist de vérification](docs/CHECKLIST_MANUELLE.md)). Sa défaite
affiche un écran de victoire.

## Brancher le back

Toutes les données du personnage sont dans `player.go`. Si la struct
de l'équipe s'appelle autrement, c'est le seul fichier à adapter :
les fichiers `*_ui.go` ne font qu'afficher ce qu'ils y trouvent. `UseItem`
(`inventory.go`) est l'endroit
où brancher `takePot` et les autres effets d'objets hors combat ;
`combat.go` (`Attack`, `UseBagItem`) fait de même pendant un combat.

## Vérification

Le projet n'embarque **pas de tests automatisés** : la validation se
fait à la main, à partir de la
[checklist de vérification manuelle](docs/CHECKLIST_MANUELLE.md).

Elle couvre 12 scénarios (création du personnage, déplacements,
inventaire, marchand, monstres, combat, poison, loot, forgeron, mort,
boss, niveaux) avec les textes attendus mot pour mot. Elle sert aussi
de référence avant/après chaque refactoring : toute différence
d'affichage est une régression.

Le détail tâche par tâche se trouve dans
[docs/TABLEAU_TACHES.md](docs/TABLEAU_TACHES.md).

## Divergences signalées

Le PDF de l'énoncé n'était pas joint à cette implémentation ; tout a
été construit à partir de la description détaillée transmise en
consigne. Voir [docs/TABLEAU_TACHES.md](docs/TABLEAU_TACHES.md#divergences-connues-à-vérifier-contre-le-pdf)
pour la liste des points (arborescence `/src` + `/docs`, numérotation
précise 1 à 22) à revérifier contre le vrai PDF avant la remise.

**Équilibrage assumé, divergent de l'énoncé** (décision d'équipe après
playtest, constantes en haut de `monster.go`) :

| Règle de l'énoncé | Valeur retenue | Raison |
|---|---|---|
| 6 monstres maximum sur la carte | **9** (`MonsterMaxAlive`) | carte de 25 salles : à 6, on croise trop rarement un monstre |
| Vie du monstre = 2x les PV max du joueur | **1x** (`MonsterHPMultiplier`) | combats plus courts et plus nerveux ; le boss reste à 2x cette valeur |
| — | vitesse monstre 70 % du joueur (`MonsterSpeedFactor`) | la fuite reste possible mais demande de ne pas hésiter |
