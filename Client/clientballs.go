package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

/*
La structure point2D
*/
type point2D struct {
	x int
	y int
	// Compteur qui sera actualisé lors de l'analyse des résultats
	c int
}

// Variable globale qui comprend tous les points du plan une fois la méthode CreatePoints() appelée
var points []point2D
var wg sync.WaitGroup

/*
Méthode permettant de stocker dans la variable globale points la totalité des positions du plan
*/

func CreatePoints() {
	for i := 1; i < 10; i++ {
		for j := 1; j < 10; j++ {

			points = append(points, point2D{i, j, 0})

		}
	}
}

const (
	IP   = "127.0.0.1" // IP local
	PORT = "1234"      // Port utilisé
)

func main() {

	// Connexion au serveur
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", IP, PORT))

	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(conn)

	message, _ := reader.ReadString('\n')

	fmt.Println(message)

	parameters, err := os.Open("parameters.txt")
	handlerErr(err)

	defer parameters.Close()

	buffy := bufio.NewScanner(parameters)

	data := ""
	for buffy.Scan() {
		data = buffy.Text()
	}
	data += "\n"

	checkParameters(data)

	fmt.Println("Envoyé au serveur :", data)
	io.WriteString(conn, data)

	wg.Add(1)

	go readConn(conn)

	wg.Wait()

	conn.Close()

}

func readConn(conn net.Conn) {
	r := bufio.NewReader(conn)

	d, err := r.ReadString('\n')

	if err != nil && err != io.EOF {
		fmt.Println("Error reading:", err.Error())
	}

	analysis(d)

	wg.Done()
}

func analysis(data string) {

	CreatePoints()
	b := strings.TrimSuffix(data, "\n")
	a := strings.Split(b, ";")

	// Création d'une slice qui contiendra les positions atteintes d'après le fichier results.csv
	positions := []point2D{}

	// Chaque objet dans data est une slice qui contient les informations d'une balle
	for _, ball := range a {

		if ball == "" {
			continue
		}

		// On récupère la position horizontale
		itemp := fmt.Sprintf("%c", ball[0])
		i, err := strconv.Atoi(itemp)
		if err != nil {
			// handle error
			fmt.Println(err)
			os.Exit(2)
		}

		// On récupère la position verticale
		jtemp := fmt.Sprintf("%c", ball[1])
		j, err := strconv.Atoi(jtemp)
		if err != nil {
			// handle error
			fmt.Println(err)
			os.Exit(2)
		}

		//Création d'un point2D "sans" compteur
		p := point2D{i, j, -1}

		positions = append(positions, p)

	}

	// Pour tous les points du plan...
	for _, point := range points {
		// On initialise un compteur à 0
		count := 0

		//Pour toutes les positions où une balle a atteint la cible
		for _, result := range positions {

			// On compare
			if point.x == result.x && point.y == result.y {

				count++

			}
		}

		point.c = count
		fmt.Printf("Au point (%d;%d), il y a eu %d bonnes balles\n", point.x, point.y, point.c)

	}

}

func checkParameters(data string) {
	params := strings.Split(data, ";")

	x, err := strconv.Atoi(params[0])
	handlerErr(err)

	if reflect.TypeOf(x) != reflect.TypeOf(reflect.Int) {
		fmt.Println("La position X n'est pas valide")
		os.Exit(1)
	}
}

func handlerErr(err error) {
	if err != nil {
		panic(err)
	}
}
