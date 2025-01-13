package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// Fonction qui prend une fonction en argument et mesure le temps d'exécution et la charge CPU
func mesurerPerformance(maFonction func()) {
	// Afficher le nombre de CPUs disponibles
	numCPU := runtime.NumCPU()
	fmt.Printf("Nombre de CPU disponibles : %d\n", numCPU)

	// Mesurer l'utilisation du CPU avant l'exécution
	start := time.Now()
	beforeCPU := runtime.MemStats{}
	runtime.ReadMemStats(&beforeCPU)

	// Appeler la fonction passée en paramètre
	maFonction()

	// Mesurer l'utilisation du CPU après l'exécution
	afterCPU := runtime.MemStats{}
	runtime.ReadMemStats(&afterCPU)

	// Calculer la durée
	duration := time.Since(start)

	// Afficher les informations sur le temps d'exécution et l'utilisation du CPU
	fmt.Printf("Temps d'exécution : %v\n", duration)
	fmt.Printf("Mémoire avant : %v\n", beforeCPU.Sys)
	fmt.Printf("Mémoire après : %v\n", afterCPU.Sys)
}

// Fonction pour compter de 0 à a et afficher les valeurs via des goroutines
func compter(a int) {
	var wg sync.WaitGroup // Utiliser un WaitGroup pour attendre la fin des goroutines

	// Boucle pour créer des goroutines
	for i := 0; i < a; i++ {
		wg.Add(1) // Ajouter une goroutine à attendre
		go func(i int) {
			defer wg.Done() // Signaler la fin de la goroutine
			fmt.Println(i)
		}(i)
	}

	// Attendre que toutes les goroutines se terminent
	wg.Wait()
}

func compter2(a int) {

	// Boucle pour créer des goroutines
	for i := 0; i < a; i++ {
		fmt.Println(i)

	}

}
func main() {
	// Passer une fonction anonyme qui appelle compter(10) à mesurerPerformance
	mesurerPerformance(func() {
		compter(1000)
	})
}
