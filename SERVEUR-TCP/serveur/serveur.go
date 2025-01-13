package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/jpeg" // Important pour décoder JPEG
	_ "image/png"  // Pour PNG, au cas où
	"io"
	"math"
	"net/http"
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
	// Configure le handler pour l'endpoint /upload
	http.HandleFunc("/upload", handleUpload)

	// Démarre le serveur sur le port 8080
	port := "8080"
	fmt.Printf("Serveur démarré sur http://localhost:%s\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("Erreur lors du démarrage du serveur : %v\n", err)
	}
}

// handleUpload gère la réception de l'image
func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée, utilisez POST", http.StatusMethodNotAllowed)
		return
	}

	// Parse la requête multipart pour récupérer les fichiers
	err := r.ParseMultipartForm(10 << 20) // Limite à 10 Mo
	if err != nil {
		http.Error(w, "Erreur lors de la lecture des données", http.StatusBadRequest)
		return
	}

	// Récupère le fichier envoyé avec le champ "image"
	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Erreur lors de la récupération du fichier", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fmt.Printf("Fichier reçu : %s (%d octets)\n", handler.Filename, handler.Size)

	// Enregistre l'image sur le disque
	savePath := "./images/" + handler.Filename
	err = os.MkdirAll("./images", os.ModePerm)
	if err != nil {
		http.Error(w, "Erreur lors de la création du dossier", http.StatusInternalServerError)
		return
	}
	outFile, err := os.Create(savePath)
	if err != nil {
		http.Error(w, "Erreur lors de la sauvegarde du fichier", http.StatusInternalServerError)
		return
	}
	defer outFile.Close()

	// Copie les données du fichier reçu dans le fichier local
	_, err = io.Copy(outFile, file)
	if err != nil {
		http.Error(w, "Erreur lors de l'écriture du fichier", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Fichier sauvegardé : %s\n", savePath)

	// Maintenant, chargeons l'image et appliquons le flou gaussien
	// Réinitialisation du fichier à sa position initiale
	file.Seek(0, io.SeekStart)

	// Décodons l'image depuis le fichier
	img, _, err := image.Decode(file)
	if err != nil {
		http.Error(w, "Erreur lors de la lecture de l'image", http.StatusInternalServerError)
		return
	}

	// Définissons la taille du noyau et le sigma pour le flou
	kernelSize := 5 // Par exemple, un noyau 5x5
	sigma := 10.0   // Paramètre sigma pour le flou

	// Générer le noyau gaussien
	kernel := generateGaussianKernel(kernelSize, sigma)

	// Créer une nouvelle image RGBA pour stocker l'image floutée
	bounds := img.Bounds()
	blurred := image.NewRGBA(bounds)

	// Application du flou gaussien sur l'image
	var wg sync.WaitGroup
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		wg.Add(1)
		go applyGaussianBlurToRow(img, kernel, y, bounds, blurred, &wg)
	}

	// Attendre que toutes les goroutines finissent
	wg.Wait()

	// Sauvegarder l'image floutée dans un nouveau fichier
	blurredPath := "./images/blurrred_" + handler.Filename
	outFileBlurred, err := os.Create(blurredPath)
	if err != nil {
		http.Error(w, "Erreur lors de la sauvegarde de l'image floutée", http.StatusInternalServerError)
		return
	}
	defer outFileBlurred.Close()

	// Sauvegarder l'image floutée en format JPEG
	err = jpeg.Encode(outFileBlurred, blurred, nil)
	if err != nil {
		http.Error(w, "Erreur lors de la sauvegarde de l'image floutée", http.StatusInternalServerError)
		return
	}

	// Répondre au client avec le chemin du fichier flouté
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Image floutée reçue et sauvegardée sous %s", blurredPath)))

}
