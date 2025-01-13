package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

func main() {
	// Chemin vers l'image à envoyer
	imagePath := "./test.jpg"

	// Ouvre le fichier image
	file, err := os.Open(imagePath)
	if err != nil {
		fmt.Printf("Erreur lors de l'ouverture du fichier : %v\n", err)
		return
	}
	defer file.Close()

	// Prépare la requête multipart
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Ajoute le fichier au champ "image"
	part, err := writer.CreateFormFile("image", "test.jpg")
	if err != nil {
		fmt.Printf("Erreur lors de la création du champ multipart : %v\n", err)
		return
	}
	_, err = io.Copy(part, file)
	if err != nil {
		fmt.Printf("Erreur lors de la copie du fichier : %v\n", err)
		return
	}
	writer.Close()

	// Envoie la requête POST au serveur
	url := "http://localhost:8080/upload"
	req, err := http.NewRequest(http.MethodPost, url, &requestBody)
	if err != nil {
		fmt.Printf("Erreur lors de la création de la requête : %v\n", err)
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Erreur lors de l'envoi de la requête : %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Affiche la réponse du serveur
	fmt.Printf("Statut : %s\n", resp.Status)
	fmt.Println("Réponse du serveur :")
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	fmt.Println(buf.String())
}
