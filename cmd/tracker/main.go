package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"
	"uuid"

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
	case "update":
		handleUpdate(s, args[1:])
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
	fmt.Println("  add    		Ajouter une candidature")
	fmt.Println("  list   		Lister les candidatures")
	fmt.Println("  update 		Mettre à jour le status d'une candidature")
	fmt.Println("\nOptions globales :")
	fmt.Println("  --data  		Chemin vers le fichier JSON (default: ~/.candidature-tracker/data.json)")
	fmt.Println("\nOptions pour 'add' : ")
	fmt.Println("  --company	Nom de l'entreprise")
	fmt.Println("  --position	Intitulé du poste")
	fmt.Println("  --url		Lien de l'offre")
	fmt.Println("  --notes		Notes")
	fmt.Println("\nOptions pour 'update' : ")
	fmt.Println(" --id			ID de la candidature (requis)")
	fmt.Println(" --status		Nouveau statut (requis)")
}

// handleAdd gère la sous-commande "add".
func handleAdd(s store.Store, args []string) {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	company := fs.String("company", "", "Nom de l'entreprise (requis)")
	position := fs.String("position", "", "Intitulé du poste (requis)")
	link := fs.String("url", "", "Lien de l'offre (optionnel)")
	notes := fs.String("notes", "", "Notes (optionnel)")

	fs.Parse(args)

	c, err := model.New(*company, *position, *link, *notes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur de validation : %v\n", err)
		os.Exit(1)
	}

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

	// Initialisation du writer
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "ID\tENTREPRISE\tPOSTE\tSTATUT\n")

	for _, c := range list {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", c.ID, c.Company, c.Position, c.Status)
	}
	w.Flush()
}

// handleUpdate gère la sous-commande "update".
func handleUpdate(s store.Store, args []string) {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	idStr := fs.String("id", "", "ID de la candidature (requis)")
	statusStr := fs.String("status", "", "Nouveau statut (requis)")
	fs.Parse(args)

	if *idStr == "" || *statusStr == "" {
		fmt.Fprintln(os.Stderr, "Erreur : les options --id et --status sont requises")
		os.Exit(1)
	}

	// 1. Convertir la chaîne en UUID
	id, err := uuid.Parse(*idStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur : ID invalide (%v)\n", err)
		os.Exit(1)
	}

	// 2. Récupérer la candidature
	c, err := s.Get(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			fmt.Fprintf(os.Stderr, "Erreur : aucune candidature trouvée avec l'ID %s\n", id)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Erreur de lecture : %v\n", err)
		os.Exit(1)
	}

	// 3. Mettre à jour le statut et la date
	c.Status = model.Status(*statusStr)
	c.UpdatedAt = time.Now()

	// 4. Valider (au cas où l'utilisateur a tapé un statut invalide)
	if err := c.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Erreur de validation : %v\n", err)
		os.Exit(1)
	}

	// 5. Sauvegarder
	if err := s.Save(c); err != nil {
		fmt.Fprintf(os.Stderr, "Erreur de sauvegarde : %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Statut mis à jour avec succès !\nID: %s\nEntreprise: %s\nNouveau statut: %s\n", c.ID, c.Company, c.Status)
}
