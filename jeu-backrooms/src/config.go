package main

// ---------------------------------------------------------------
// CONFIGURATION — tous les réglages chiffrés du moteur, au même endroit.
// Les réglages d'équilibrage propres à un domaine (dégâts, prix…)
// restent à côté de leurs règles (combat.go, economy.go…).
// ---------------------------------------------------------------

// ---------------------------------------------------------------
// RÉGLAGES — tout ce que tu peux vouloir changer est ici
// ---------------------------------------------------------------
const (
	assetsDir  = "assets"       // dossier des ressources
	startScene = "level1_start" // scène de départ

	TileSize = 32 // taille d'une case en pixels
	GridW    = 25 // nombre de cases en largeur
	GridH    = 15 // nombre de cases en hauteur

	ScreenW = TileSize * GridW // 800
	ScreenH = TileSize * GridH // 480

	startFullscreen = false // true pour démarrer directement en plein écran

	playerSpeed      = 2.0 // pixels par frame
	animSpeed        = 8   // ticks entre deux images d'animation
	playerFrames     = 4   // images d'animation par direction dans player.png
	playerDirections = 4   // lignes de player.png : bas, gauche, droite, haut

	maxNameLength = 14  // lettres maximum dans le nom du personnage
	toastDuration = 120 // durée d'affichage d'un message temporaire (2 s)

	LevelRows = 5 // carte de niveau 1 : 5x5 salles
	LevelCols = 5

	// Salle de départ (alias "level1_start") et salle du boss, en haut au
	// milieu de la carte (tâche boss).
	startRoomName = "level1_r3c2"
	bossRoomName  = "level1_r0c2"
)

// Boîte de collision du joueur (plus petite que le sprite : les "pieds").
const (
	boxOffX = 8
	boxOffY = 18
	boxW    = 16
	boxH    = 12
)

// Directions — correspondent aux LIGNES du spritesheet.
const (
	DirDown  = 0
	DirLeft  = 1
	DirRight = 2
	DirUp    = 3
)

var dirName = map[int]string{DirDown: "sud", DirLeft: "ouest", DirRight: "est", DirUp: "nord"}

const (
	MonsterXPReward   = 20  // mission bonus expérience : XP gagnée par monstre normal
	BossXPReward      = 100 // XP gagnée en battant le boss
	NPCInteractRadius = 44.0

	// Après avoir refusé le combat contre le boss, délai avant que son
	// dialogue puisse se rouvrir (le temps de s'éloigner).
	bossDialogGrace = MonsterTransitionGrace * 2
)
