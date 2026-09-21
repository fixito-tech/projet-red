#!/usr/bin/env python3
"""
Génère assets/player.png : spritesheet du personnage du Projet RED.
Combinaison antiradiation jaune + masque à gaz noir (thème Backrooms).

Disposition attendue par le moteur :
  4 colonnes (images d'animation) x 4 lignes (bas, gauche, droite, haut)
Taille d'une image : 40 x 48 px  ->  feuille de 160 x 192 px
"""
from PIL import Image, ImageDraw

FW, FH = 40, 48           # taille d'une image
COLS, ROWS = 4, 4

# --- palette ---
SUIT       = (232, 201, 58)    # jaune combinaison
SUIT_DARK  = (186, 158, 38)    # plis / ombre
SUIT_LIGHT = (246, 226, 122)   # reflet
RUBBER     = (28, 28, 26)      # masque, gants, bottes
RUBBER_LT  = (58, 58, 54)      # reliefs du caoutchouc
LENS       = (46, 58, 48)      # verre du masque
LENS_HI    = (146, 168, 148)   # reflet sur le verre
TANK       = (92, 92, 88)      # bonbonne dorsale
STRAP      = (44, 44, 40)      # sangles
LINE       = (26, 24, 14)      # contour

# balancement des jambes/bras sur les 4 images
SWING = [0, 3, 0, -3]
# petit rebond du corps pendant la marche
BOB   = [0, -1, 0, -1]


def box(d, x0, y0, x1, y1, fill, outline=None):
    d.rectangle([x0, y0, x1, y1], fill=fill, outline=outline)


# ----------------------------------------------------------------
# TÊTE — la capuche jaune reste bien visible autour du masque noir
# ----------------------------------------------------------------
def head_front(d, cx, top):
    box(d, cx - 7, top, cx + 6, top + 16, SUIT, LINE)           # capuche
    box(d, cx - 6, top + 1, cx + 5, top + 2, SUIT_LIGHT)        # reflet
    box(d, cx - 5, top + 6, cx + 4, top + 13, RUBBER)           # masque
    for ox in (-4, 1):                                           # hublots
        box(d, cx + ox, top + 8, cx + ox + 2, top + 10, LENS)
        d.point((cx + ox, top + 8), fill=LENS_HI)
    box(d, cx - 2, top + 13, cx + 1, top + 16, RUBBER)          # cartouche
    box(d, cx - 1, top + 14, cx, top + 15, RUBBER_LT)


def head_back(d, cx, top):
    box(d, cx - 7, top, cx + 6, top + 16, SUIT, LINE)
    box(d, cx - 6, top + 1, cx + 5, top + 2, SUIT_LIGHT)
    box(d, cx - 6, top + 7, cx + 5, top + 8, STRAP)             # sangles du masque
    box(d, cx - 6, top + 11, cx + 5, top + 12, STRAP)
    box(d, cx - 2, top + 7, cx, top + 12, RUBBER)               # boucle


def head_side(d, cx, top):
    box(d, cx - 6, top, cx + 6, top + 16, SUIT, LINE)
    box(d, cx - 5, top + 1, cx + 5, top + 2, SUIT_LIGHT)
    box(d, cx + 1, top + 6, cx + 6, top + 13, RUBBER)           # masque, moitié avant
    box(d, cx + 2, top + 8, cx + 4, top + 10, LENS)             # hublot
    d.point((cx + 2, top + 8), fill=LENS_HI)
    box(d, cx + 6, top + 10, cx + 8, top + 13, RUBBER)          # cartouche en avant
    box(d, cx - 6, top + 7, cx - 3, top + 9, STRAP)             # sangle arrière


# ----------------------------------------------------------------
# TORSE
# ----------------------------------------------------------------
def body_front(d, cx, top, swing, back=False):
    box(d, cx - 8, top, cx + 7, top + 13, SUIT, LINE)
    box(d, cx - 7, top + 1, cx + 6, top + 2, SUIT_LIGHT)
    if back:
        box(d, cx - 4, top + 3, cx + 3, top + 10, TANK)         # bonbonne dorsale
        box(d, cx - 4, top + 5, cx + 3, top + 6, RUBBER_LT)
        box(d, cx - 4, top + 8, cx + 3, top + 9, RUBBER_LT)
    else:
        box(d, cx - 1, top + 3, cx, top + 9, SUIT_DARK)         # fermeture éclair
        box(d, cx + 3, top + 4, cx + 6, top + 7, SUIT_DARK)     # poche
    box(d, cx - 7, top + 10, cx + 6, top + 11, RUBBER)          # ceinture
    for sign in (-1, 1):
        s = swing if sign > 0 else -swing
        x0 = cx - 11 if sign < 0 else cx + 8
        y0 = top + 2 + max(0, s)
        box(d, x0, y0, x0 + 2, y0 + 8, SUIT, LINE)              # bras
        box(d, x0, y0 + 8, x0 + 2, y0 + 10, RUBBER)             # gant


def body_side(d, cx, top, swing):
    box(d, cx - 6, top, cx + 5, top + 13, SUIT, LINE)
    box(d, cx - 5, top + 1, cx + 4, top + 2, SUIT_LIGHT)
    box(d, cx - 6, top + 3, cx - 4, top + 9, SUIT_DARK)         # pli dorsal
    box(d, cx - 6, top + 10, cx + 4, top + 11, RUBBER)          # ceinture
    x0 = cx + 1 + swing                                          # bras avant
    box(d, x0, top + 3, x0 + 2, top + 9, SUIT_DARK, LINE)
    box(d, x0, top + 9, x0 + 2, top + 11, RUBBER)


# ----------------------------------------------------------------
# JAMBES
# ----------------------------------------------------------------
def legs_front(d, cx, top, f):
    # sur les images 1 et 3, une jambe se lève (elle est plus courte)
    lift = {0: (0, 0), 1: (2, 0), 2: (0, 0), 3: (0, 2)}[f]
    for sign, up in ((-1, lift[0]), (1, lift[1])):
        x0 = cx - 7 if sign < 0 else cx + 2
        x1 = x0 + 4
        box(d, x0, top, x1, top + 7 - up, SUIT, LINE)
        box(d, x0, top + 7 - up, x1, top + 9 - up, RUBBER, LINE)


def legs_side(d, cx, top, swing):
    # jambe arrière (plus sombre), puis jambe avant par-dessus
    for s, col in ((-swing, SUIT_DARK), (swing, SUIT)):
        x0 = cx - 3 + s
        x1 = x0 + 5
        box(d, x0, top, x1, top + 7, col, LINE)
        box(d, x0 - 1, top + 7, x1, top + 9, RUBBER, LINE)


def draw_frame(img, col, row, kind, f):
    d = ImageDraw.Draw(img)
    ox, oy = col * FW, row * FH
    sub = Image.new("RGBA", (FW, FH), (0, 0, 0, 0))
    ds = ImageDraw.Draw(sub)

    cx = FW // 2
    bob = BOB[f]
    swing = SWING[f]
    head_top = 4 + bob
    body_top = 21 + bob
    legs_top = 35 + bob

    if kind == "front":
        legs_front(ds, cx, legs_top, f)
        body_front(ds, cx, body_top, swing)
        head_front(ds, cx, head_top)
    elif kind == "back":
        legs_front(ds, cx, legs_top, f)
        body_front(ds, cx, body_top, swing, back=True)
        head_back(ds, cx, head_top)
    else:  # profil droit
        legs_side(ds, cx, legs_top, swing)
        body_side(ds, cx, body_top, swing)
        head_side(ds, cx, head_top)

    img.paste(sub, (ox, oy), sub)


def main():
    sheet = Image.new("RGBA", (FW * COLS, FH * ROWS), (0, 0, 0, 0))
    # ligne 0 : bas (face) / 1 : gauche / 2 : droite / 3 : haut (dos)
    for f in range(COLS):
        draw_frame(sheet, f, 0, "front", f)
        draw_frame(sheet, f, 2, "side", f)
        draw_frame(sheet, f, 3, "back", f)
    # la ligne "gauche" est le miroir de la ligne "droite"
    right = sheet.crop((0, 2 * FH, FW * COLS, 3 * FH))
    left = Image.new("RGBA", (FW * COLS, FH), (0, 0, 0, 0))
    for f in range(COLS):
        cell = right.crop((f * FW, 0, (f + 1) * FW, FH))
        left.paste(cell.transpose(Image.FLIP_LEFT_RIGHT), (f * FW, 0))
    sheet.paste(left, (0, 1 * FH))

    sheet.save("assets/player.png")
    print("assets/player.png :", sheet.size)

    # aperçu agrandi pour contrôle visuel
    sheet.resize((sheet.width * 4, sheet.height * 4), Image.NEAREST).save("/tmp/player_preview.png")


if __name__ == "__main__":
    main()
