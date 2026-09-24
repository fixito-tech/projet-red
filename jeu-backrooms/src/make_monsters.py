#!/usr/bin/env python3
"""
Génère les sprites des PNJ et monstres du Projet RED :
  assets/monster.png     — le "monstre spaghetti" (niveau 1)
  assets/boss.png         — le boss de fin de niveau (même disposition)
  assets/merchant.png     — le marchand (image fixe)
  assets/blacksmith.png   — le forgeron (image fixe)

Ces fichiers sont OPTIONNELS : si l'un d'eux est absent, le moteur Go
dessine un visuel de secours directement dans le code (voir
monster_draw.go / economy_ui.go), exactement comme pour player.png.
Lance ce script si tu veux remplacer ces visuels de secours par de
vrais sprites, générés de façon reproductible (pas de fichier binaire
à maintenir à la main).

Disposition attendue par le moteur pour monster.png / boss.png :
  4 colonnes (images de marche) x 1 ligne, ancré par les pieds.
  Taille d'une image : 56 x 72 px -> feuille de 224 x 72 px.
  Le moteur retourne le sprite horizontalement selon la direction, et
  agrandit encore le boss (bossScale, monster_draw.go).

merchant.png / blacksmith.png : une seule image fixe de 40 x 48 px
(même gabarit que le personnage), ancrée par les pieds.
"""
import math
import random

from PIL import Image, ImageDraw

FW, FH, COLS = 56, 72, 4

BODY = (10, 10, 12, 255)
BODY_BOSS = (20, 2, 4, 255)
EYE = (214, 43, 43, 255)
EYE_BOSS = (255, 40, 40, 255)


def draw_spaghetti(d, phase, dark):
    """Une silhouette faite de filaments emmêlés : un blob central et
    plusieurs "nouilles" ondulantes qui pendent vers le bas."""
    body = BODY_BOSS if dark else BODY
    eye = EYE_BOSS if dark else EYE
    cx = FW // 2

    # tête/corps : blob ovale
    d.ellipse([cx - 15, 6, cx + 15, 34], fill=body)

    # filaments : plusieurs courbes sinusoïdales, épaisseur variable
    random.seed(7) # même tirage à chaque génération : reproductible
    for strand in range(8):
        base_x = cx - 20 + strand * 6 + random.randint(-2, 2)
        width = random.choice([2, 3])
        points = []
        for y in range(24, FH - 4):
            wig = int(7 * math.sin(y * 0.22 + phase + strand * 1.3))
            points.append((base_x + wig, y))
        for i in range(len(points) - 1):
            d.line([points[i], points[i + 1]], fill=body, width=width)

    # deux yeux rouges, seul signe de vie dans la masse noire
    d.ellipse([cx - 7, 16, cx - 4, 19], fill=eye)
    d.ellipse([cx + 4, 16, cx + 7, 19], fill=eye)


def make_creature_sheet(path, dark):
    sheet = Image.new("RGBA", (FW * COLS, FH), (0, 0, 0, 0))
    for f in range(COLS):
        frame = Image.new("RGBA", (FW, FH), (0, 0, 0, 0))
        d = ImageDraw.Draw(frame)
        draw_spaghetti(d, phase=f * 1.6, dark=dark)
        sheet.paste(frame, (f * FW, 0), frame)
    sheet.save(path)
    print(path, ":", sheet.size)


# ----------------------------------------------------------------
# PNJ — silhouettes fixes simples (marchand / forgeron)
# ----------------------------------------------------------------
SKIN = (227, 189, 146, 255)
LINE = (26, 24, 14, 255)


def make_npc(path, coat, accent, prop):
    img = Image.new("RGBA", (40, 48), (0, 0, 0, 0))
    d = ImageDraw.Draw(img)
    d.rectangle([13, 4, 26, 13], fill=SKIN, outline=LINE)      # tête
    d.rectangle([8, 14, 31, 43], fill=coat, outline=LINE)      # manteau/tablier
    d.rectangle([19, 18, 20, 29], fill=accent)                 # fermeture / lanière
    if prop == "bag":
        d.rectangle([28, 24, 36, 34], fill=accent, outline=LINE)   # sacoche du marchand
    elif prop == "hammer":
        d.rectangle([30, 10, 33, 26], fill=(120, 84, 40, 255))     # manche
        d.rectangle([26, 6, 37, 12], fill=(90, 90, 94, 255), outline=LINE)  # tête de marteau
    img.save(path)
    print(path, ":", img.size)


def main():
    make_creature_sheet("assets/monster.png", dark=False)
    make_creature_sheet("assets/boss.png", dark=True)
    make_npc("assets/merchant.png", coat=(106, 74, 42, 255), accent=(216, 192, 122, 255), prop="bag")
    make_npc("assets/blacksmith.png", coat=(58, 58, 61, 255), accent=(176, 90, 36, 255), prop="hammer")


if __name__ == "__main__":
    main()
