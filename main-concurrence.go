package main

import (
	"image"
	"image/color"
	"image/jpeg"
	_ "image/jpeg" // Important pour décoder JPEG
	_ "image/png"  // Pour PNG, au cas où
	"log"
	"math"
	"os"
	"sync"
)

// Génère un noyau gaussien 2D
func generateGaussianKernel(size int, sigma float64) [][]float64 {
	kernel := make([][]float64, size)
	sum := 0.0
	mid := size / 2

	for i := 0; i < size; i++ {
		kernel[i] = make([]float64, size)
		for j := 0; j < size; j++ {
			x, y := float64(-i+mid), float64(-j+mid)
			value := (1 / (2 * math.Pi * sigma * sigma)) * math.Exp(-(x*x+y*y)/(2*sigma*sigma))
			kernel[i][j] = value
			sum += value
		}
	}

	// Normalisation
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			kernel[i][j] /= sum
		}
	}

	return kernel
}

// Applique un noyau gaussien à une ligne de pixels
func applyGaussianBlurToRow(img image.Image, kernel [][]float64, y int, bounds image.Rectangle, blurred *image.RGBA, wg *sync.WaitGroup) {
	defer wg.Done() // Indiquer que cette goroutine est terminée

	kSize := len(kernel)
	kMid := kSize / 2

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		var rSum, gSum, bSum float64

		for ky := 0; ky < kSize; ky++ {
			for kx := 0; kx < kSize; kx++ {
				offsetX := x + kx - kMid
				offsetY := y + ky - kMid

				// Vérifier si les coordonnées sont valides
				if offsetX < bounds.Min.X || offsetX >= bounds.Max.X || offsetY < bounds.Min.Y || offsetY >= bounds.Max.Y {
					continue
				}

				r, g, b, _ := img.At(offsetX, offsetY).RGBA()
				weight := kernel[ky][kx]

				rSum += weight * float64(r)
				gSum += weight * float64(g)
				bSum += weight * float64(b)
			}
		}

		// Définir la nouvelle couleur du pixel
		blurred.Set(x, y, color.RGBA{
			R: uint8(rSum / 256),
			G: uint8(gSum / 256),
			B: uint8(bSum / 256),
			A: 255,
		})
	}
}

func main() {
	// Ouvre l'image d'entrée
	file, err := os.Open("images/newYork.jpg")
	if err != nil {
		log.Fatalf("Erreur ouverture fichier : %v", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatalf("Erreur décodage image : %v", err)
	}

	// Génère un noyau gaussien large
	kernel := generateGaussianKernel(10, 5.0) // Taille 15x15, sigma = 5.0

	bounds := img.Bounds()
	blurred := image.NewRGBA(bounds)

	// Création d'un WaitGroup pour attendre que toutes les goroutines finissent
	var wg sync.WaitGroup

	// Lancer une goroutine pour chaque ligne de pixels
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		wg.Add(1)
		go applyGaussianBlurToRow(img, kernel, y, bounds, blurred, &wg)
	}

	// Attendre que toutes les goroutines aient terminé
	wg.Wait()

	// Sauvegarde l'image floutée
	outFile, err := os.Create("images/newYorkApresFlou-c-10-5.jpg")
	if err != nil {
		log.Fatalf("Erreur création fichier sortie : %v", err)
	}
	defer outFile.Close()

	err = jpeg.Encode(outFile, blurred, nil)
	if err != nil {
		log.Fatalf("Erreur encodage image sortie : %v", err)
	}

	log.Println("Flou gaussien appliqué et image sauvegardée dans output.jpg")
}
