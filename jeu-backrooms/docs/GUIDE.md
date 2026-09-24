# Guide de compréhension du code

Ce guide explique comment le jeu est organisé et comment une partie se
déroule, fonction par fonction. Tous les fichiers sont dans `src/`, dans
le même `package main` : Go les voit comme un seul programme, on les a
seulement rangés par thème.

**Une règle à retenir :** chaque écran a un fichier `*_input.go` (ce que
font les touches) et un fichier `*_ui.go` (ce qu'on dessine). Les règles
du jeu (PV, prix, dégâts…) sont dans les fichiers sans suffixe et ne
dessinent jamais rien.

---

## 1. Le plan du code

### Le cœur
| Fichier | Rôle |
|---|---|
| `main.go` | Point d'entrée : charge la salle de départ, prépare le jeu et lance la boucle d'Ebiten. |
| `game.go` | La struct `Game` (tout l'état de la partie) et les méthodes `Update` / `Draw` qu'Ebiten appelle 60 fois par seconde. |
| `config.go` | Les réglages chiffrés du moteur : taille des cases, vitesse du joueur, noms des salles… |
| `input.go` | Petites fonctions qui disent si une touche de menu vient d'être appuyée. |

### Les règles du jeu (aucun dessin)
| Fichier | Rôle |
|---|---|
| `player.go` | La struct `Character`, les 3 classes, la création, l'XP, la mort, le soin. |
| `inventory.go` | L'inventaire : ajouter, retirer, agrandir, utiliser un objet. |
| `equipment.go` | La struct `Equipment` (tête, torse, pieds) et le bonus de PV max. |
| `spells.go` | Les sorts appris une seule fois, et les attaques verrouillées. |
| `economy.go` | Le marchand (acheter, vendre) et le forgeron (recettes, fabrication). |
| `combat.go` | Le combat au tour par tour : attaques, sac, fuite, riposte, poison. |
| `monster.go` | La struct `Monster`, l'apparition sur la carte, l'IA, le loot. |
| `boss.go` | Le boss et son aura, qui le rend imbattable sans équipement. |
| `scene.go` | Les salles : lecture des fichiers, murs, portes, arrivée dans la salle suivante. |
| `world.go` | Ce qui vit tout seul pendant l'exploration : rencontres, réapparitions, loot, délais. |
| `errors.go` | Les messages d'erreur personnalisés (pas assez de pièces, inventaire plein…). |

### Les écrans (clavier + dessin)
| Fichiers | Écran |
|---|---|
| `menu_input.go`, `menu_ui.go` | Menu titre, saisie du nom, choix de la classe, victoire (clavier). |
| `play_input.go`, `play_ui.go` | Exploration : déplacements, portes, inventaire ouvert, touches de test. |
| `inventory_ui.go` | Le panneau d'inventaire et les icônes des objets. |
| `economy_input.go`, `economy_ui.go` | Les menus du marchand et du forgeron. |
| `combat_input.go`, `combat_ui.go` | L'écran de combat, le dialogue du boss, l'écran de victoire. |
| `monster_draw.go` | Les monstres et les objets posés au sol. |
| `ui.go` | Les outils de dessin partagés (texte, panneaux, jauges), le HUD, le message temporaire. |
| `placeholders.go` | Les visuels dessinés dans le code quand une image PNG manque. |

---

## 2. Le déroulement d'une partie

**Un mot important : les « images ».** Ebiten appelle `Update` puis `Draw`
60 fois par seconde. Tous les délais du jeu se comptent donc en images
(120 images = 2 secondes). On n'utilise jamais `time.Sleep` : ça figerait
l'écran.

1. **Démarrage** — `main` charge la salle de départ (`LoadScene`), crée le
   `Game` sur l'écran `StateMenu`, charge les images (`loadImages`),
   règle la fenêtre (`setupWindow`) puis lance `ebiten.RunGame`.

2. **À chaque image** — `Game.Update` fait un `switch` sur `g.state` et
   appelle la fonction de l'écran affiché (`updateMenu`, `updatePlay`,
   `updateCombat`…). `Game.Draw` fait le même `switch` pour dessiner, puis
   dessine le message temporaire par-dessus.

3. **Menu titre** — `updateMenu` : « Jouer » passe à `StateCreate`.

4. **Création du personnage** (`menu_input.go`)
   - `updateNameEntry` garde seulement les lettres ; `FormatNom` met la
     1re en majuscule et le reste en minuscules.
   - `updateClassChoice` : à la validation, `NewCharacter` crée le
     personnage (niveau 1, moitié de ses PV max, 3 Almond Water, 30 pièces)
     puis `startNewGame` prépare la carte.

5. **Nouvelle partie** (`world.go`) — `startNewGame` pose le joueur sur le
   `@` de la salle de départ (`placeAtStart`), puis `populateWorld` fait
   apparaître 9 monstres (`SpawnMonster`), le boss (`NewBoss`), le
   marchand et le forgeron (`PlaceNPCs`).

6. **Exploration** — `updatePlay`, dans cet ordre :
   1. touches de test F2 à F4 (`updateDebugKeys`), délais de grâce,
      énergie qui remonte, monstres qui réapparaissent ;
   2. `handlePlayKeys` : inventaire (E), parler à un PNJ (R), menu (Échap) ;
   3. `movePlayer` : déplacement axe par axe, pour glisser le long des murs ;
   4. `checkTransition` : sur une porte `+`, `goTo` charge la salle suivante ;
   5. `updateDropPickup` : ramasser le loot en marchant dessus ;
   6. `updateEncounters` : un monstre qui touche le joueur lance le combat
      (`startCombat`) ; toucher le boss ouvre le dialogue de confirmation.

7. **Marchand et forgeron** — `openNPC` ouvre `StateShop` ou `StateCraft`.
   `updateShop` achète (`Character.Buy`) ou vend (`Character.Sell`),
   `updateCraft` fabrique (`Character.Craft`), qui équipe aussitôt la pièce
   (`EquipItem`).

8. **Combat** — voir la partie 4.

9. **Fin de combat** — `endCombat` :
   - victoire : XP gagnée (`AddXP`), loot posé au sol (`dropLoot`), monstre
     retiré ; si c'était le boss, écran de victoire ;
   - défaite : `Die` remet la moitié des PV max, `teleportToStart` renvoie
     à la salle de départ ;
   - dans tous les cas (sauf le boss vaincu), retour à l'exploration avec
     un court délai sans nouveau combat.

10. **Victoire** — `drawVictory` affiche l'écran de fin ; Entrée ramène au
    menu titre.

---

## 3. Les structs principales

### `Character` (player.go) — le personnage
| Champ | À quoi il sert |
|---|---|
| `Nom`, `Classe`, `Niveau` | Identité du personnage. |
| `PV`, `PVMax` | Points de vie actuels et maximum. |
| `Energie`, `EnergieMax` | Le « mana » : payé par les attaques spéciales. |
| `Inventaire` | La liste des objets (des noms, ex. `"Almond Water"`). |
| `Capacite`, `Ameliorations` | Places dans l'inventaire (10, +10 par amélioration, 3 fois max). |
| `Pieces` | L'argent, à part de l'inventaire. |
| `Equip` | L'équipement porté (struct `Equipment`). |
| `SpellsKnown` | Les sorts appris (un sort ne s'apprend qu'une fois). |
| `Vitesse` | L'initiative : le plus rapide frappe en premier. |
| `XP`, `XPMax` | Expérience actuelle et nécessaire pour le niveau suivant. |
| `BasePVMax`, `BaseEnergieMax`, `NiveauBonusPV`, `NiveauBonusEN` | Valeurs de départ et bonus de niveau, gardés à part pour recalculer les maximums (`RecomputeMaxStats`). |

### `Equipment` (equipment.go) — l'équipement
Trois emplacements, `Head`, `Chest` et `Feet`, chacun vide (`nil`) ou
occupé par un `EquipmentItem` (nom, emplacement, bonus de PV max).
`Set` équipe une pièce et renvoie l'ancienne, qui retourne dans
l'inventaire.

### `Monster` (monster.go) — un monstre
| Champ | À quoi il sert |
|---|---|
| `Room`, `X`, `Y` | Salle et position. |
| `PV`, `PVMax` | Points de vie (égaux aux PV max de base du joueur, x2 pour le boss). |
| `AI`, `Seen` | Errance ou poursuite ; `Seen` affiche le « ! ». |
| `Dir`, `Frame`, `Tick` | Direction et animation de marche. |
| `Alive`, `IsBoss`, `Vitesse` | Vivant ou non, boss ou non, initiative. |

### `Combat` (combat.go) — un combat en cours
| Champ | À quoi il sert |
|---|---|
| `Player`, `Monster` | Les deux adversaires. |
| `Round` | Le compteur de tours. |
| `Result` | En cours, victoire, défaite ou fuite. |
| `Messages` | La file des phrases à afficher une par une. |
| `PoisonTurnsLeft` | Tours de poison restants sur le monstre. |
| `ShakeTimer`, `FlashTimer` | Effets visuels quand un coup porte. |
| `FleeBlocked` | Vrai contre le boss : impossible de fuir. |

### `Game` (game.go) — toute la partie
Elle regroupe le personnage (`hero`), la salle affichée (`scene`), la
position du joueur (`px`, `py`), l'écran affiché (`state`), les monstres,
le loot, les PNJ, le combat en cours et les sélections des menus. Chaque
champ est commenté dans `game.go`.

---

## 4. Un tour de combat, expliqué simplement

Imagine un jeu de cartes où chacun joue à son tour.

1. **Début du tour** (`startRound`) : le compteur de tours augmente. Si le
   monstre est plus rapide que le joueur (champ `Vitesse`), il frappe tout
   de suite.
2. **Le joueur choisit** dans le menu Attaque / Sac / Fuite :
   - **Attaque** (`Combat.Attack`) : on vérifie que le sort est appris, que
     le soin est utile et qu'il reste assez d'énergie. Sinon un message
     s'affiche et le joueur rechoisit, sans perdre son tour. Si tout va
     bien, on paie l'énergie et l'attaque fait son effet
     (`applyAttackEffect`) : des dégâts au monstre, ou un soin.
   - **Sac** (`Combat.UseBagItem`) : l'Almond Water soigne 50 PV, la potion
     de poison empoisonne le monstre pour 4 tours.
   - **Fuite** (`Combat.Flee`) : le combat s'arrête, sauf contre le boss.
3. **Après l'action du joueur** (`afterPlayerAction`) : si le monstre n'a
   plus de PV, c'est gagné. Sinon, s'il n'a pas encore joué, il riposte
   (`monsterAttack`).
4. **La riposte** (`monsterAttack`) : le monstre fait entre 6 et 12 dégâts
   (14 à 20 pour le boss). **Tous les 3 tours, ses dégâts sont doublés**
   (`Round % 3 == 0`, message « motif renforcé »). Le boss ajoute son
   aura : 60 dégâts, moins 20 par pièce d'équipement portée.
5. **Fin du tour** (`endRound`) : le poison retire 8 PV au monstre, le
   joueur récupère 12 d'énergie, puis un nouveau tour commence.

Chaque étape ajoute une phrase dans `Messages`. L'écran les affiche une
par une, et le joueur ne peut rien faire tant qu'il en reste
(`HasMessages`).

**Exemple :** le Survivant (vitesse 10) contre un monstre normal
(vitesse 8). Le joueur étant plus rapide, il attaque d'abord à chaque
tour, puis le monstre riposte. Au 3e tour, la riposte fait double dégâts.

---

## 5. Questions qu'un jury pourrait poser

**Pourquoi tous les fichiers sont-ils dans le même package ?**
Ebiten impose que `Update` et `Draw` soient des méthodes de `Game`, et
presque tous les écrans lisent `Game`. Un seul package évite des dizaines
de majuscules et des imports croisés : on a rangé le code par fichiers,
sans compliquer.

**Comment le jeu sait-il quel écran afficher ?**
Grâce au champ `state` de `Game`. `Update` et `Draw` font un `switch`
dessus et appellent la fonction de l'écran : `StatePlay` donne
`updatePlay`, `StateCombat` donne `updateCombat`, etc.

**Pourquoi des compteurs d'images plutôt que `time.Sleep` ?**
Le jeu doit redessiner l'écran 60 fois par seconde. `time.Sleep`
bloquerait tout le programme et figerait l'image. On compte donc les
images : 90 images = 1,5 seconde.

**Comment l'équipement augmente-t-il les PV max ?**
`RecomputeMaxStats` recalcule toujours `PVMax = PV de base + bonus de
niveau + bonus d'équipement`. Comme la base est gardée à part, on ne
perd jamais la valeur d'origine quand on change de pièce.

**Comment empêche-t-on d'apprendre deux fois le même sort ?**
`LearnSpell` vérifie avec `KnowsSpell` ; si le sort est déjà connu, elle
renvoie l'erreur `ErrSpellAlreadyKnown` (« Ce sort est déjà appris. »).

**Comment fonctionnent les erreurs personnalisées ?**
`GameError` est un simple texte qui implémente l'interface `error` de Go
(méthode `Error`). Chaque erreur porte directement la phrase à afficher.

**Comment le monstre double-t-il ses dégâts tous les 3 tours ?**
Dans `monsterAttack` : si `Round % 3 == 0`, les dégâts sont multipliés
par 2 et le message dit « motif renforcé ».

**Comment l'inventaire est-il limité ?**
`AddItem` refuse l'objet si `len(Inventaire) >= Capacite`.
`UpgradeInventory` ajoute 10 places, au plus 3 fois.

**Pourquoi le jeu garde-t-il les salles en mémoire (`sceneCache`) ?**
Charger une salle, c'est lire un fichier et décoder une image. Comme une
salle ne change jamais, on la charge une seule fois et on la réutilise à
chaque passage.

**Que se passe-t-il si une image manque ?**
`loadSprite` renvoie alors un visuel de secours dessiné dans le code
(`placeholders.go`) : le jeu tourne même sans les images.
