# Projet RED — Backrooms

Mini jeu de rôle en ligne de commande (CLI), écrit en Go, dans lequel vous incarnez
un survivant qui vient de "no-clip" dans les **Backrooms**. Explorez, gérez votre
inventaire, échangez des ressources avec le Troqueur, fabriquez de l'équipement
chez le Bricoleur, et affrontez les entités errantes en combat tour par tour.

Projet réalisé dans le cadre du **Projet RED** (Ynov Campus Aix).

## Fonctionnalités

- Création d'un survivant (nom, classe : Équilibré / Éclaireur / Robuste)
- Fiche personnage, inventaire, équipement (tête, torse, pieds)
- Le Troqueur : achat de potions, matériaux et améliorations
- Le Bricoleur : fabrication d'équipement à partir de matériaux récoltés
- Combat tour par tour avec initiative, sorts, mana (Adrénaline) et expérience
- Système de niveau et de montée en puissance

## Prérequis

- [Go](https://go.dev/dl/) 1.21 ou supérieur

## Installation

```bash
git clone https://github.com/fixito-tech/projet-red.git
cd projet-red/src
```

## Lancement du jeu

Depuis le dossier `src` :

```bash
go run .
```

Ou en compilant un exécutable :

```bash
go build -o projet-red .
./projet-red
```

## Tests

La logique du jeu (personnage, inventaire, marchand, fabrication, combat...) est
couverte par des tests automatisés, indépendants des menus interactifs. Depuis le
dossier `src` :

```bash
go test ./...       # résumé
go test -v ./...     # détail de chaque test
```

## Structure du projet

```
.
├── docs/   # Document de gestion de projet
├── src/    # Code source du jeu (Go)
└── README.md
```

## Auteurs

- Antoine Mortelette
