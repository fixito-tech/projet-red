package main

// ---------------------------------------------------------------
// VISUELS DE SECOURS — dessinés directement dans le code, pixel par
// pixel, pour que le jeu tourne même quand une image PNG manque
// (personnage, monstres, PNJ, décor d'une salle).
// ---------------------------------------------------------------

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Couleurs des visuels de secours.
var (
	colPlaceholderBody = color.RGBA{0x2b, 0x4a, 0x8b, 0xff} // joueur
	colPlaceholderSkin = color.RGBA{0xf0, 0xcb, 0xa0, 0xff}
	colMonsterBody     = color.RGBA{0x0a, 0x0a, 0x0c, 0xff}
	colBossBody        = color.RGBA{0x14, 0x02, 0x04, 0xff}
	colMonsterEye      = color.RGBA{0xd6, 0x2b, 0x2b, 0xff}
	colNPCSkin         = color.RGBA{0xe3, 0xbd, 0x92, 0xff}
	colFallbackFloor   = color.RGBA{0xc9, 0xbb, 0x3d, 0xff} // décor
	colFallbackWall    = color.RGBA{0xa8, 0x99, 0x2a, 0xff}
	colFallbackDoor    = color.RGBA{0x6b, 0x5e, 0x12, 0xff}
)

// Couleurs des PNJ : manteau + accessoire (sacoche, marteau).
var (
	colMerchantCoat   = color.RGBA{0x6a, 0x4a, 0x2a, 0xff}
	colMerchantAccent = color.RGBA{0xd8, 0xc0, 0x7a, 0xff}
	colSmithCoat      = color.RGBA{0x3a, 0x3a, 0x3d, 0xff}
	colSmithAccent    = color.RGBA{0xb0, 0x5a, 0x24, 0xff}
)

// placeholderSheet dessine un personnage de secours très simple, utilisé
// si assets/player.png est absent : même disposition 4 x 4 que le vrai.
func placeholderSheet() *ebiten.Image {
	const fw, fh, cols, rows = 32, 32, 4, 4
	img := image.NewRGBA(image.Rect(0, 0, fw*cols, fh*rows))
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			drawPlaceholderPlayer(img, col*fw, row*fh, col%2 == 1)
		}
	}
	return ebiten.NewImageFromImage(img)
}

// drawPlaceholderPlayer dessine le personnage de secours en (ox, oy) :
// tête, corps, puis jambes, écartées une image sur deux (stepping).
func drawPlaceholderPlayer(img *image.RGBA, ox, oy int, stepping bool) {
	legShift := 0
	if stepping {
		legShift = 2
	}
	paintRect(img, ox+11, oy+4, 10, 9, colPlaceholderSkin)  // tête
	paintRect(img, ox+9, oy+13, 14, 12, colPlaceholderBody) // corps
	paintRect(img, ox+10, oy+25, 4, 5-legShift, colPlaceholderBody)
	paintRect(img, ox+18, oy+25, 4, 3+legShift, colPlaceholderBody)
}

// placeholderMonsterSheet dessine une silhouette de "spaghetti" noire
// faite de filaments emmêlés : 4 images de marche (pas besoin de Pillow
// pour que le jeu tourne). dark = version plus sombre, pour le boss.
func placeholderMonsterSheet(dark bool) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, monsterFrameW*monsterCols, monsterFrameH))
	body := colMonsterBody
	if dark {
		body = colBossBody
	}
	for col := 0; col < monsterCols; col++ {
		drawSpaghetti(img, col*monsterFrameW, float64(col)*1.6, body)
	}
	return ebiten.NewImageFromImage(img)
}

// drawSpaghetti dessine une image du monstre commençant en ox : un blob
// ovale, six filaments ondulants (phase décale l'ondulation d'une image
// à l'autre) et deux yeux rouges.
func drawSpaghetti(img *image.RGBA, ox int, phase float64, body color.RGBA) {
	cx := ox + monsterFrameW/2
	// tête/corps : un blob ovale au centre-haut
	for y := 4; y < 26; y++ {
		half := int(14 - math.Abs(float64(y-14))*0.5)
		paintRect(img, cx-half, y, 2*half+1, 1, body)
	}
	// filaments : des "nouilles" de 3 pixels d'épaisseur qui pendent.
	// Bug d'origine conservé : baseX ne tient pas compte de ox, donc tous
	// les filaments tombent dans la première image de l'animation.
	for strand := 0; strand < 6; strand++ {
		baseX := monsterFrameW/2 - 18 + strand*7
		for y := 20; y < monsterFrameH-4; y++ {
			wig := int(6 * math.Sin(float64(y)*0.25+phase+float64(strand)))
			paintRect(img, baseX+wig, y, 3, 1, body)
		}
	}
	// deux yeux rouges, seul signe de vie dans la masse noire
	img.Set(cx-5, 13, colMonsterEye)
	img.Set(cx+5, 13, colMonsterEye)
}

// placeholderNPCSheet dessine un PNJ minimal (visuel de secours tant
// que make_monsters.py n'a pas été exécuté avec Pillow).
func placeholderNPCSheet(coat, accent color.RGBA) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, npcFrameW, npcFrameH))
	paintRect(img, 13, 4, 14, 10, colNPCSkin) // tête
	paintRect(img, 8, 14, 24, 30, coat)       // manteau
	paintRect(img, 19, 18, 2, 12, accent)     // sacoche ou marteau
	return ebiten.NewImageFromImage(img)
}

// fallbackBackground dessine un décor jaune "backrooms" basique depuis la grille,
// utilisé quand le PNG de la scène n'existe pas encore.
func fallbackBackground(grid [][]rune) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, ScreenW, ScreenH))
	for r := 0; r < GridH; r++ {
		for c := 0; c < GridW; c++ {
			paintRect(img, c*TileSize, r*TileSize, TileSize, TileSize, fallbackTileColor(grid[r][c]))
		}
	}
	return ebiten.NewImageFromImage(img)
}

// fallbackTileColor donne la couleur d'une case : mur, porte ou sol.
func fallbackTileColor(tile rune) color.RGBA {
	switch tile {
	case '#':
		return colFallbackWall
	case '+':
		return colFallbackDoor
	}
	return colFallbackFloor
}

// paintRect colorie un rectangle de w x h pixels dont le coin haut-gauche
// est en (x0, y0). Tous les visuels de secours sont faits de rectangles.
func paintRect(img *image.RGBA, x0, y0, w, h int, c color.RGBA) {
	for y := y0; y < y0+h; y++ {
		for x := x0; x < x0+w; x++ {
			img.Set(x, y, c)
		}
	}
}
