package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
)

/*
Remarques de PFR :
- Enlever la récursion
- Pas de variable hard codée et passer les vecteurs et le nombre de rebonds en paramètres
- Pas d'intéractif
- Gérer les erreurs
*/

// Constantes du programmes => modifiables pour changer les conditions des simulations
var sizeWall int = 10

// Variables globales
var wg sync.WaitGroup
var id = 0
var idSim = 0

/*
La structure ball permet la simulation. Elle se crée avec un identifiant,
une position initiale et un vecteur vitesse initial, ainsi qu'une valeur
représentant la vie d'une balle (qui est décrémenté à chaque percussion
avec un mur)
*/

type ball struct {
	id int
	x  int
	y  int
	vx int
	vy int

	// The initial values must be saved to be written in the results.csv file
	xmem  int
	ymem  int
	vxmem int
	vymem int

	hp int
}

/*
La fonction impactMur(b *ball) permet de gérer les impacts des balles sur les murs
La position et le nouveau vecteur sont calculés
*/

func impactMur(b *ball) {
	if b.y <= 0 { // percute mur bas
		//DEBUG
		//fmt.Println("Impact sur le mur du bas")

		b.y = b.y - b.vy // On recule d'une étape

		distance_restante := b.vy - (0 - b.y)
		b.y = 0 - distance_restante

		b.vy = -b.vy // Inversion du vecteur vitesse
		b.hp--       // La balle perd 1 hp
	}

	if b.y >= sizeWall { // percute mur haut
		//DEBUG
		//fmt.Println("Impact sur le mur du haut")

		b.y = b.y - b.vy // On recule d'une étape

		distance_restante := b.vy - (sizeWall - b.y)
		b.y = sizeWall - distance_restante

		b.vy = -b.vy // Inversion du vecteur vitesse
		b.hp--       // La balle perd 1 hp
	}

	if b.x <= 0 { // percute mur gauche
		//DEBUG
		//fmt.Println("Impact sur le mur de gauche")

		b.x = b.x - b.vx // On recule d'une étape

		distance_restante := b.vx - (0 - b.x)
		b.x = 0 - distance_restante

		b.vx = -b.vx // Inversion du vecteur vitesse
		b.hp--       // La balle perd 1 hp
	}

	if b.x >= sizeWall { // percute mur droit
		//DEBUG
		//fmt.Println("Impact sur le mur de droite")

		b.x = b.x - b.vx // On recule d'une étape

		distance_restante := b.vx - (sizeWall - b.x)
		b.x = sizeWall - distance_restante

		b.vx = -b.vx // Inversion du vecteur vitesse
		b.hp--       // La balle perd 1 hp
	}
}

/*
La fonction (b *ball) actualizePosition() est une fonction de l'objet ball
C'est une fonction récursive qui permet de simuler le déplacement de la balle dans le plan
*/

func (b *ball) actualizePosition(targetX int, targetY int) {

	// Variable de sortie de récursion
	fin := false

	// On avance une fois
	b.x = b.x + b.vx
	b.y = b.y + b.vy

	// On vérifie s'il y a eu un impact. Si oui la positio est actualisée dans la méthode impactMur(b)
	impactMur(b)

	//DEBUG
	// fmt.Printf("La balle %d est en position %d, %d | Il reste %d hp\n", b.id, b.x, b.y, b.hp)

	// On vérifie si la balle a atteint la cible. Si oui, fin deviendra true
	fin = targetReached(b, targetX, targetY)

	// On vérifie si la balle n'est pas morte (plus de 10 rebonds)
	if b.hp <= 0 {
		//DEBUG
		// fmt.Printf("La balle %d est morte en position %d, %d\n", b.id, b.x, b.y)
		// On arrête la simulation
		fin = true
	}

	if !fin {
		// Récursion
		b.actualizePosition(targetX, targetY)

	} else {
		// Cette simulation a terminée, on peut en notifier le waiting group
		wg.Done()

	}

}

/*
La fonction targetReached(b *ball) permet de vérifier si la balle a atteint la
cible établie au début du programme. Elle renvoie un boolean true si c'est le cas
*/

func targetReached(b *ball, targetX int, targetY int) bool {

	reached := false

	if b.x == targetX && b.y == targetY {
		//DEBUG
		//fmt.Printf("Target reached at position %d, %d | %d HP remaining\n", b.x, b.y, b.hp)

		addToFile(b, targetX, targetY) // Ajoute au fichier les données de la ball b
		reached = true                 // On renvoie true ce qui va stoppé la récursion
	}
	return reached
}

/*
La fonction addToFile(b *ball) permet d'écrire dans le fichier results.csv les données
de la balle si elle a atteint la cible. Les données envoyées sont l'id, la position initiale
et le vecteur initial stockés dans l'objet
! A noter : seule la position initale est véritablement nécessaire pour notre utilisation des données
mais une étude différente pourrait porter sur les vecteurs initiaux.
L'identifiant est transmis pour pouvoir comparer lors du DEBUG
*/

func addToFile(b *ball, targetX int, targetY int) {
	// Création d'une slice contenant les données convertit sous forme de strings
	data := []string{strconv.Itoa(b.id), strconv.Itoa(b.xmem), strconv.Itoa(b.ymem), strconv.Itoa(b.vxmem), strconv.Itoa(b.vymem)}

	file := fmt.Sprintf("result%d%d%d.csv", targetX, targetY, idSim)

	CSVfile, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)

	if err != nil {
		log.Fatal(err)
	}

	// Fermera le fichier plus tard
	defer CSVfile.Close()

	writer := csv.NewWriter(CSVfile)

	writer.Write(data)
	writer.Flush()

	//DEBUG
	// fmt.Printf("Data of ball %d added to the file\n", b.id)
}

/*
La fonction launchBall(i int, j int) permet d'envoyer dans des goroutines
toutes les simulations pour une position initiale (i,j) donnée.
12 vecteurs sont utilisés (disponible en bas du programme)
*/

func lauchBall(i int, j int, targetX int, targetY int) {

	for _, vector := range vectors {
		// Il faut incrémenter l'id pour chaque balle créée
		id++

		// Nouvelle balle prête à être lancée (on tolère jusqu'à 10 rebonds compris)
		b := ball{id, i, j, vector[0], vector[1], i, j, vector[0], vector[1], 10}

		//DEBUG
		// fmt.Printf("Ball id : %d |Ball x : %d |Ball y : %d |Ball vx : %d |Ball vy : %d |Ball hp : %d\n", b.id, b.x, b.y, b.vx, b.vy, b.hp)

		go b.actualizePosition(targetX, targetY)

	}

}

/*
La fonction main du programme lance la simulation
*/

func main() {

	// Mise en route du serveur
	ln, err := net.Listen("tcp", ":1234")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	defer ln.Close()

	fmt.Println("Listening on port 1234")

	for {

		conn, err := ln.Accept()

		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}

		go handleRequest(conn)

	}

}

func handleRequest(conn net.Conn) {

	fmt.Printf("***Connexion au client\n\n")

	io.WriteString(conn, "Connexion au serveur établie\n")

	reader := bufio.NewReader(conn)

	data, err := reader.ReadString('\n')

	if err != nil && err != io.EOF {
		fmt.Println("Error reading:", err.Error())
	}

	fmt.Printf("La cible a atteindre pour cette simulation est le point (%c;%c)\n\n", data[0], data[1])

	i, _ := strconv.Atoi(string(data[0]))
	j, _ := strconv.Atoi(string(data[1]))

	dataInt := []int{i, j}

	startSimulation(dataInt, conn)
}

func startSimulation(data []int, conn net.Conn) {
	targetX := data[0]
	targetY := data[1]

	idSim++
	fmt.Printf("Starting simulation %d\n", idSim)

	file := fmt.Sprintf("result%d%d%d.csv", targetX, targetY, idSim)
	os.Create(file)

	// Lance 81 goroutines : 1 par position du plan, sans compter les bords (0 ou 10)
	for i := 1; i < sizeWall; i++ {
		for j := 1; j < sizeWall; j++ {
			// Chacune de ces goroutines va lancer 12 autres goroutines, il faut donc ajouter 12 au waiting group
			wg.Add(12)
			go lauchBall(i, j, targetX, targetY)
		}
	}

	wg.Wait()

	// Sends data on the connexion

	f, err := os.Open(file)

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	validBalls, err := csv.NewReader(f).ReadAll()

	if err != nil {
		log.Fatal(err)
	}

	str := ""

	for _, ball := range validBalls {
		s := fmt.Sprintf("%s%s;", ball[1], ball[2])
		str += s
	}
	str += "\n"
	fmt.Println(str)
	io.WriteString(conn, str)

}

// Tous les vecteurs différents testés pour chaque position du plan
// On a exclu les vecteurs résultants dans des solutions évidentes (comme les vecteurs horizontaux et verticaux)
var vectors = [][]int{
	{1, 1},
	{1, -1},
	{-1, 1},
	{-1, -1},
	{1, 2},
	{1, -2},
	{-1, 2},
	{-1, -2},
	{2, 1},
	{2, -1},
	{-2, 1},
	{-2, -1},
}
