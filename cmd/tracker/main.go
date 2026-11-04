package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mickael-thenot/candidature-tracker/internal/model"
	"github.com/mickael-thenot/candidature-tracker/internal/store"
)

func main() {
	// 1. Flag global pour le chemin du fichier de données
	dataPath := flag.String("data", defaultDataPath(), "Chemin vers le fichier de données JSON")
	flag.Parse()

	// 2. Initialisation du store
	s, err := store.NewJSONStore(*dataPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur d'initialisation : %v\n", err)
		os.Exit(1)
	}

	// 3. Récupération de la sous-commande
	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	cmd := args[0]
	switch cmd {
	case "add":
		handleAdd(s, args[1:])
	case "list":
		handleList(s)
	default:
		fmt.Fprintf(os.Stderr, "Commande inconnue : %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func defaultDataPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur pour trouver le dossier utilisateur : %v\n", err)
		os.Exit(1)
	}
	return filepath.Join(home, ".candidature-tracker", "data.json")
}

func printUsage() {
	fmt.Println("Usage: tracker <commande> [options]")
	fmt.Println("Commandes :")
	fmt.Println("  add    Ajouter une candidature")
	fmt.Println("  list   Lister les candidatures")
	fmt.Println("\nOptions globales :")
	fmt.Println("  --data  Chemin vers le fichier JSON (default: ~/.candidature-tracker/data.json)")
	fmt.Println("\nOptions pour 'add' : ")
	fmt.Println("  --company	Nom de l'entreprise")
	fmt.Println("  --position	Intitulé du poste")
	fmt.Println("  --url		Lien de l'offre")
	fmt.Println("  --notes		Notes")
}

// handleAdd gère la sous-commande "add".
func handleAdd(s store.Store, args []string) {
	// On utilise un FlagSet local pour ne pas interférer avec les flags globaux (comme --data)
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	company := fs.String("company", "", "Nom de l'entreprise (requis)")
	position := fs.String("position", "", "Intitulé du poste (requis)")
	link := fs.String("url", "", "Lien de l'offre (optionnel)")
	notes := fs.String("notes", "", "Notes (optionnel)")

	fs.Parse(args)

	// Création de la candidature (la validation se fait dans model.New)
	c, err := model.New(*company, *position, *link, *notes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur de validation : %v\n", err)
		os.Exit(1)
	}

	// Sauvegarde
	if err := s.Save(c); err != nil {
		fmt.Fprintf(os.Stderr, "Erreur de sauvegarde : %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Candidature ajoutée avec succès !\nID: %s\nEntreprise: %s\nPoste: %s\n", c.ID, c.Company, c.Position)
}

// handleList gère la sous-commande "list".
func handleList(s store.Store) {
	list, err := s.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur de lecture : %v\n", err)
		os.Exit(1)
	}

	if len(list) == 0 {
		fmt.Println("Aucune candidature trouvée.")
		return
	}

	fmt.Printf("%-38s | %-20s | %-20s | %-10s\n", "ID", "Entreprise", "Poste", "Statut")
	fmt.Println("-------------------------------------------------------------------------------------------------")
	for _, c := range list {
		fmt.Printf("%-38s | %-20s | %-20s | %-10s\n", c.ID, c.Company, c.Position, c.Status)
	}
}
