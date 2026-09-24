package main

// ---------------------------------------------------------------
// LE JEU — l'état de la partie (struct Game), les trois méthodes
// qu'Ebiten appelle en boucle (Update, Draw, Layout), les messages
// temporaires et le chargement des images au démarrage.
// ---------------------------------------------------------------

import (
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// ---------------------------------------------------------------
// ÉTAT DU JEU
// ---------------------------------------------------------------

// State dit quel écran est affiché : Update et Draw font un switch dessus.
type State int

const (
	StateMenu        State = iota
	StateCreate            // saisie du nom puis choix de la classe
	StatePlay              // exploration des salles
	StateCombat            // combat au tour par tour façon Pokémon
	StateShop              // menu du marchand
	StateCraft             // menu du forgeron
	StateBossConfirm       // dialogue de confirmation avant d'affronter le boss
	StateVictory           // écran de victoire (boss vaincu)
)

// Game contient tout l'état de la partie. Les délais et compteurs sont
// en images : le jeu en affiche 60 par seconde.
type Game struct {
	state State // écran affiché

	// Le joueur dans la salle
	hero      *Character // statistiques, inventaire, équipement (player.go)
	scene     *Scene     // salle affichée
	px, py    float64    // position du joueur (coin haut-gauche du sprite)
	dir       int        // direction regardée (DirDown…) = ligne du spritesheet
	walkFrame int        // image de l'animation de marche (0 à 3)
	walkTick  int        // images écoulées depuis le dernier pas d'animation

	// Images du joueur et du menu
	sheet          *ebiten.Image // spritesheet du joueur : 4 images x 4 directions
	frameW, frameH int           // taille d'une image du spritesheet
	menuBg         *ebiten.Image // fond du menu titre (nil s'il n'y en a pas)

	// Menus et interface
	menuIndex      int      // choix sélectionné dans le menu titre
	createStep     int      // création du personnage : 0 = nom, 1 = classe
	nameBuf        []rune   // nom en cours de saisie
	classSel       int      // classe sélectionnée
	invOpen        bool     // inventaire ouvert (touche E)
	invSel         int      // case sélectionnée dans l'inventaire
	npcMenuSel     int      // ligne sélectionnée chez le marchand ou le forgeron
	shopMode       ShopMode // onglet du marchand : acheter ou vendre
	bossConfirmSel int      // dialogue du boss : 0 = Oui, 1 = Non
	toastMsg       string   // message temporaire affiché en bas de l'écran
	toastTimer     int      // images restantes avant qu'il disparaisse
	uiTick         int      // compteur d'images pour les animations des menus
	showDebug      bool     // infos de débogage (F1)

	// Le monde : monstres, loot, PNJ, boss
	rng                  *rand.Rand // hasard de la carte (apparitions, loot)
	monsters             []*Monster
	drops                []*Drop // objets posés au sol
	monsterRespawnTimer  int     // images avant la prochaine réapparition
	merchant, blacksmith *NPC
	boss                 *Monster
	bossKilled           bool

	// Délais de grâce : pendant ce temps, pas de nouvelle rencontre
	roomGrace   int // après un changement de salle
	combatGrace int // après un combat
	bossGrace   int // avant que le dialogue du boss puisse se rouvrir
	energyTick  int // compteur de la régénération d'énergie hors combat

	// Le combat en cours
	combat           *Combat
	combatMenu       CombatMenu // menu ouvert : principal, attaques ou sac
	combatSel        int        // choix sélectionné dans ce menu
	combatEnterTimer int        // images restantes du fondu d'entrée
}

// ---------------------------------------------------------------
// LES TROIS MÉTHODES APPELÉES PAR EBITEN — 60 fois par seconde, Ebiten
// appelle Update (faire avancer le jeu) puis Draw (dessiner l'écran).
// ---------------------------------------------------------------

// Update fait avancer le jeu d'une image, selon l'écran affiché.
func (g *Game) Update() error {
	// Plein écran : F11 ou Alt+Entrée, disponible dans le menu comme en jeu.
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) ||
		(inpututil.IsKeyJustPressed(ebiten.KeyEnter) && ebiten.IsKeyPressed(ebiten.KeyAlt)) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}

	g.uiTick++
	g.tickToast()

	switch g.state {
	case StateMenu:
		return g.updateMenu()
	case StateCreate:
		return g.updateCreate()
	case StatePlay:
		return g.updatePlay()
	case StateCombat:
		return g.updateCombat()
	case StateShop:
		return g.updateShop()
	case StateCraft:
		return g.updateCraft()
	case StateBossConfirm:
		return g.updateBossConfirm()
	case StateVictory:
		return g.updateVictory()
	}
	return nil
}

// Draw dessine l'écran correspondant à l'état du jeu. Les fonctions de
// dessin elles-mêmes sont dans les fichiers *_ui.go.
func (g *Game) Draw(screen *ebiten.Image) {
	switch g.state {
	case StateMenu:
		g.drawMenu(screen)
	case StateCreate:
		drawCreate(screen, g)
	case StatePlay:
		g.drawPlay(screen)
	case StateCombat:
		drawCombat(screen, g)
	case StateShop:
		g.drawPlay(screen)
		drawMerchantMenu(screen, g)
	case StateCraft:
		g.drawPlay(screen)
		drawBlacksmithMenu(screen, g)
	case StateBossConfirm:
		g.drawPlay(screen)
		drawBossConfirm(screen, g)
	case StateVictory:
		drawVictory(screen, g)
	}

	// Le message temporaire est dessiné en tout dernier, par-dessus
	// n'importe quel écran : sinon il restait invisible en combat et
	// caché derrière les menus du marchand et du forgeron.
	drawToast(screen, g.toastMsg)
}

// Layout donne à Ebiten la résolution interne du jeu : 800x480,
// agrandie ensuite à la taille de la fenêtre.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenW, ScreenH
}

// ---------------------------------------------------------------
// MESSAGES TEMPORAIRES (ex. "Inventaire plein !")
// ---------------------------------------------------------------

// toast affiche un message en bas de l'écran pendant 2 secondes.
func (g *Game) toast(m string) {
	g.toastMsg = m
	g.toastTimer = toastDuration
}

// tickToast fait disparaître le message temporaire une fois son délai écoulé.
func (g *Game) tickToast() {
	if g.toastTimer > 0 {
		g.toastTimer--
		if g.toastTimer == 0 {
			g.toastMsg = ""
		}
	}
}

// loadImages charge le sprite du joueur (ou dessine celui de secours)
// et l'image de fond du menu, si elle existe.
func (g *Game) loadImages() {
	// Spritesheet : 4 colonnes (animation) x 4 lignes (bas, gauche, droite, haut).
	g.sheet = loadSprite("player.png", placeholderSheet())
	g.frameW = g.sheet.Bounds().Dx() / playerFrames
	g.frameH = g.sheet.Bounds().Dy() / playerDirections

	// Image de fond du menu : optionnelle, nil si le fichier manque.
	g.menuBg = loadSprite("menu.png", nil)
}

// setupWindow règle la fenêtre : deux fois la résolution interne (plus
// confortable, et l'image reste nette), redimensionnable, et en plein
// écran dès le départ si startFullscreen vaut true.
func setupWindow() {
	ebiten.SetWindowSize(ScreenW*2, ScreenH*2)
	ebiten.SetWindowTitle("Projet RED — The Backrooms")
	// Autorise le redimensionnement : sans ça, le bouton "agrandir" de la
	// barre de titre est grisé.
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if startFullscreen {
		ebiten.SetFullscreen(true)
	}
}
