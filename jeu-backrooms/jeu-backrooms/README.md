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

## Le personnage

`assets/player.png` — combinaison antiradiation jaune, masque à gaz noir.
Feuille de 160x192 px : 4 colonnes (animation de marche) x 4 lignes.

| Ligne | Direction | Ce qu'on voit |
|-------|-----------|----------------|
| 0 | bas   | face : masque, deux hublots, cartouche filtrante |
| 1 | gauche| profil : un hublot, cartouche vers l'avant |
| 2 | droite| profil inversé |
| 3 | haut  | dos : sangles du masque et bonbonne dorsale |

Le sprite fait 40x48 px alors qu'une case en fait 32x32 : il déborde vers le
haut et sur les côtés, mais son empreinte au sol reste d'une case. Les
collisions ne changent donc pas.

Pour le modifier : `python3 make_player.py` (nécessite Pillow). Les couleurs
sont regroupées en haut du fichier.

## Affichage

- La fenêtre s'ouvre en 1600x960 (le jeu tourne en interne en 800x480, affiché
  en x2 : l'image reste nette).
- **F11** ou **Alt+Entrée** : bascule plein écran / fenêtre, dans le menu comme
  en jeu.
- Le bouton "agrandir" de la barre de titre fonctionne désormais
  (`SetWindowResizingMode`).
- Pour démarrer directement en plein écran : mettre `startFullscreen = true`
  en haut de `main.go`.

En plein écran, l'image est agrandie en gardant ses proportions : des bandes
noires apparaissent sur les côtés si l'écran est en 16/9. C'est voulu — ça
évite de déformer le décor.

## Interface (front)

Après « Jouer » : saisie du nom (lettres uniquement, majuscule ajoutée
automatiquement — tâche 11), puis choix de la classe.

| Classe | Base (énoncé) | PV max | Énergie max |
|--------|---------------|--------|-------------|
| Survivant | Humain | 100 | 100 |
| Ancien Résident | Elfe | 80 | 120 |
| Chasseur de niveaux | Nain | 120 | 80 |

On démarre niveau 1 avec la moitié de ses PV max et 3 Almond Water.

En jeu : barres de PV et d'énergie en haut à gauche. **E** ouvre/ferme
l'inventaire (10 emplacements, tâche 12 ; 40 max après 3 améliorations,
tâche 18). Flèches pour choisir, Entrée pour utiliser.

### Touches de test (pratiques pour la démo à l'oral)
- **F2** : ramasse un objet (montre la limite de 10)
- **F3** : prend 15 dégâts et perd 10 d'énergie
- **F4** : agrandit l'inventaire de 10 places
- **F1** : infos de débogage

### Brancher le back
Toutes les données du personnage sont dans `player.go`. Si la struct de
l'équipe s'appelle autrement, c'est le seul fichier à adapter : `ui.go` ne
fait qu'afficher ce qu'il y trouve. La fonction `UseItem` est l'endroit où
appeler `takePot` et les autres effets d'objets.

### Tests
`go test .` vérifie : format du nom, stats des classes, limite de 10 objets,
3 améliorations max, soin de l'Almond Water, transitions entre salles.
