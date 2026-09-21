package main

type Character struct {
    Name      string
    Class     string
    Level     int
    MaxHP     int
    CurrentHP int
    Inventory [10]string
}

func InitCharacter(name string, class string) Character {
    niveau := 1
    var inventaire [10]string
    var pointsVieMax int
    var pvActuel int

    if class == "leger" {
        pointsVieMax = 80
    }
    if class == "moyen" {
        pointsVieMax = 100
    }
    if class == "lourd" {
        pointsVieMax = 120
    }

    pvActuel = pointsVieMax / 2

    return Character{
        Name:      name,
        Class:     class,
        Level:     niveau,
        MaxHP:     pointsVieMax,
        CurrentHP: pvActuel,
        Inventory: inventaire,
    }
}

func displayInfo(name string, class string ) string{
	perso := InitCharacter(name string, class string )

}
func inventory() {

	var invent [10]int

	for i := 0; i < len(invent); i++ {
		if invent[i] == 0 {
			invent[i] = drop()
			break
		}
	}
}