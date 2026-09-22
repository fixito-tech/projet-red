# Tableau de correspondance — tâches de l'énoncé

⚠️ **Le PDF de l'énoncé n'a pas été fourni pour cette implémentation.**
Le développement s'est basé sur la description détaillée transmise en
consigne (parties « personnage et inventaire », « économie et
artisanat », « combat au tour par tour », + missions bonus). Les
numéros de tâche ci-dessous ne sont donc **certains** que lorsqu'ils
étaient déjà présents dans le code fourni (commentaires `tâche N`
d'origine dans `player.go`) ; partout ailleurs, la colonne « # PDF »
indique `?` et doit être vérifiée par l'équipe contre le vrai PDF avant
la soutenance. Le regroupement en 3 parties + bonus, lui, suit
fidèlement la consigne.

## Partie 1 — Personnage et inventaire

| # PDF | Tâche | Fichier | Fonction / type | Test |
|---|---|---|---|---|
| 1 | `struct Character` | [player.go](../player.go) | `Character` | `TestStatsDeClasse` |
| 11 | Création : nom (lettres uniquement, formaté), 3 classes, départ à 50 % des PV | [player.go](../player.go), [main.go](../main.go), [ui.go](../ui.go) | `NewCharacter`, `FormatNom`, `updateCreate`, `drawClassStep` | `TestFormatNom`, `TestStatsDeClasse` |
| 5 | Potion de vie (Almond Water, +50 PV) | [player.go](../player.go) | `UseItem` | `TestAlmondWater` |
| ? | Potion de poison : dégâts dans le temps, PV affichés | [combat.go](../combat.go) | `UseBagItem`, `endRound` (poison) | `TestPoisonPotionDamagesMonsterOverTime` |
| ? | Mort puis résurrection à 50 % des PV | [player.go](../player.go), [main.go](../main.go) | `Character.Die`, `endCombat` (cas `ResultLose`) | `TestDieResurrectsAtHalfPV` |
| ? | Livre de sort appris une seule fois | [spells.go](../spells.go), [economy.go](../economy.go) | `LearnSpell`, `Buy` | `TestLearnSpellOnlyOnce`, `TestBuySpellbookOnceOnly` |
| 12 | Inventaire : 10 emplacements au départ | [player.go](../player.go) | `AddItem`, `Capacite` | `TestLimiteInventaire` |
| 18 | Inventaire : +10 emplacements, 3 fois maximum | [player.go](../player.go) | `UpgradeInventory` | `TestAmeliorationInventaire` |
| ? | Argent de départ | [player.go](../player.go) | `StartingCoins`, `NewCharacter` | `TestStartingCoins` |

## Partie 2 — Économie et artisanat

| # PDF | Tâche | Fichier | Fonction / type | Test |
|---|---|---|---|---|
| ? | Marchand : achète le loot, vend potions/grimoire | [economy.go](../economy.go), [economy_ui.go](../economy_ui.go) | `Buy`, `Sell`, `MerchantSells`, `MerchantBuyPrices` | `TestBuyNotEnoughCoins`, `TestSellLootAndRejectOthers` |
| ? | Forgeron : recettes (matériaux + pièces) | [economy.go](../economy.go) | `Craft`, `Recipes` | `TestCraftMissingMaterials`, `TestCraftSuccessEquipsAndConsumes` |
| ? | `struct Equipment` (tête, torse, pieds), bonus PV max, échange avec l'ancien objet | [equipment.go](../equipment.go), [player.go](../player.go) | `Equipment`, `Equipment.Set`, `Character.EquipItem` | `TestEquipmentBonusPV`, `TestEquipmentExchange`, `TestCharacterEquipItemUpdatesMaxPVAndReturnsOld` |
| ? | Interaction avec un PNJ (touche dédiée + bulle d'aide) | [main.go](../main.go), [economy_ui.go](../economy_ui.go) | `nearbyNPC`, `openNPC`, `drawInteractionBubble` | vérifié manuellement (jeu, pas de logique pure isolable) |
| ? | Erreurs personnalisées (pièces, inventaire plein, matériaux, sort déjà connu…) | [errors.go](../errors.go) | `GameError`, `ErrNotEnoughCoins`, etc. | `TestBuyNotEnoughCoins`, `TestBuyInventoryFull`, `TestCraftMissingMaterials`, `TestBuySpellbookOnceOnly` |

## Partie 3 — Combat au tour par tour

| # PDF | Tâche | Fichier | Fonction / type | Test |
|---|---|---|---|---|
| ? | `struct Monster` | [monster.go](../monster.go) | `Monster`, `NewMonster` | `TestMonsterHPFormula` |
| ? | Pattern d'attaque : dégâts doublés tous les 3 tours | [combat.go](../combat.go) | `monsterAttack` | `TestMonsterAttackPatternDoublesEveryThirdTurn` |
| ? | Tour du joueur : attaquer / inventaire | [combat.go](../combat.go), [main.go](../main.go) | `Combat.Attack`, `Combat.UseBagItem`, `updateCombat` | `TestAttackPunchIsFree`, `TestAttackAgainstLockedSpellFails` |
| ? | Boucle de combat avec compteur de tours | [combat.go](../combat.go) | `Combat.Round`, `startRound`, `endRound` | `TestMonsterAttackPatternDoublesEveryThirdTurn` (dépend du compteur) |
| ? | Monstres : apparition, IA (errance/poursuite), contact → combat, réapparition | [monster.go](../monster.go), [main.go](../main.go) | `SpawnMonster`, `Monster.Update`, `updateMonsterRespawn` | `TestMonsterChasesWithinRadius`, `TestSpawnableRoomsExcludesStartAndBoss`, `TestIsFreeFloorRejectsNearDoors` |
| ? | Loot : drop au sol (max 3 types), ramassage | [monster.go](../monster.go), [main.go](../main.go) | `RollLoot`, `Drop`, `updateDropPickup` | `TestRollLootMaxThreeTypes` |
| ? | Écran de combat façon Pokémon (transition, barres animées, boîte de texte) | [combat_ui.go](../combat_ui.go) | `drawCombat`, `drawCombatBar` | vérifié manuellement (rendu graphique) |
| ? | Boss : placement, confirmation, imbattable sans équipement, écran de victoire | [boss.go](../boss.go), [main.go](../main.go) | `NewBoss`, `BossAuraDamage`, `updateBossConfirm`, `endCombat` | `TestBossUnbeatableWithoutEquipment`, `TestBossBeatableWithFullEquipment` |

## Missions bonus

La consigne annonce 6 missions bonus mais n'en nomme explicitement que
4 dans la description transmise ; les 2 autres sont donc absentes de
ce tableau faute d'information (à compléter contre le PDF).

| Mission bonus | Fichier | Fonction / type | Test |
|---|---|---|---|
| Initiative (le plus rapide agit en premier) | [player.go](../player.go), [combat.go](../combat.go) | `Character.Vitesse`, `Monster.Vitesse`, `startRound` | démontré par `TestMonsterAttackPatternDoublesEveryThirdTurn` (ordre des messages) |
| Expérience (XP, niveaux) | [player.go](../player.go) | `Character.AddXP` | `TestAddXPLevelsUp` |
| Sorts (grimoire verrouillé) | [spells.go](../spells.go) | `LearnSpell`, `AttackLocked` | `TestFireballLockedUntilLearned`, `TestAttackLockedOnlyForSpellAttacks` |
| Mana / énergie (coût, régénération en et hors combat) | [combat.go](../combat.go), [main.go](../main.go) | `EnergyRegenPerTurn`, `updateOutOfCombatRegen` | couvert indirectement par les tests de combat |

## Divergences connues à vérifier contre le PDF

- **Arborescence `/src` + `/docs`** : le dépôt fourni est resté à plat
  (fichiers `.go` à la racine du module), tel que reçu. Le
  réorganiser en `/src` casserait les chemins relatifs vers
  `assets/` sans bénéfice fonctionnel ; ce point est signalé plutôt
  que corrigé silencieusement — à confirmer avec l'équipe/le prof
  avant la remise si l'arborescence imposée est strictement notée.
- **Numérotation précise 1 à 22** : seules les tâches 1, 5, 11, 12 et
  18 étaient déjà annotées dans le code d'origine ; toutes les autres
  correspondances (`?`) sont des hypothèses de bonne foi à recaler sur
  le vrai PDF.
- **Valeurs chiffrées** (prix, dégâts, PV, délais) : toutes regroupées
  en constantes nommées (voir `combat.go`, `monster.go`, `boss.go`,
  `economy.go`, `player.go`) pour être ajustées en un seul endroit une
  fois comparées à l'énoncé réel.
