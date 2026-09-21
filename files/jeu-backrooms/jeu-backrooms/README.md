# Projet RED — The Backrooms (moteur graphique)

Moteur 2D en Go avec Ebiten : menu, sprite animé, collisions, transitions de scènes.

## Lancer

```bash
go mod tidy
go run .
```

## Commandes
- Menu : flèches ou ZQSD, Entrée pour valider
- Jeu : flèches ou ZQSD pour se déplacer, Échap pour revenir au menu, F1 pour afficher les infos de scène

## Ajouter une scène
Chaque scène = 2 fichiers dans `assets/scenes/` :
- `nom.png` : le décor (800x480)
- `nom.txt` : la grille de collisions 25x15 + les sorties

Caractères de la grille : `#` mur, `.` sol, `+` porte, `@` départ.
En-tête du .txt : `# scene: nom` puis `# exit: est -> autre_scene`.

## Remplacer le personnage
Dépose `assets/player.png` : un spritesheet de 4 colonnes (animation)
x 4 lignes (bas, gauche, droite, haut). Sans ce fichier, un sprite
de secours est généré automatiquement.

Un fond de menu optionnel peut être mis dans `assets/menu.png`.
