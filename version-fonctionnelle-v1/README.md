# Version fonctionnelle (v1)

Copie de sauvegarde du projet **Projet RED : Les Backrooms** dans un état qui **fonctionne**
(`go vet` et `go test ./...` passent, le jeu se lance et se joue).

Cette version est **à améliorer** : elle sert de point de repli. En cas de problème sur le
code principal (`src/`), on peut revenir à cette version.

## Contenu
- `src/` : code source du jeu (Go), avec les tâches 1 à 22, les missions bonus, un système de quêtes,
  un affichage en couleur et une carte explorable au clavier.
- `docs/` : document de gestion de projet.
- `README-projet.md` : README du projet à la date de la sauvegarde.

## Lancer cette version
```bash
cd version-fonctionnelle-v1/src
go run .
```

## Tester
```bash
cd version-fonctionnelle-v1/src
go test ./...
```

## Pistes d'amélioration
- Faire disparaître un Hurleur de la carte après sa défaite.
- Mode « touche directe » pour la carte sous Linux et macOS (actuellement Windows uniquement).
- Fusionner proprement le `character.go` de la racine avec `src/character.go`.
