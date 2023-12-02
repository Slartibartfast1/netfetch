package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/deggja/netfetch/backend/statik"

	"github.com/deggja/netfetch/backend/pkg/k8s"
	"github.com/rakyll/statik/fs"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
)

var dashCmd = &cobra.Command{
	Use:   "dash",
	Short: "Launch the Netfetch interactive dashboard",
	Run: func(cmd *cobra.Command, args []string) {
		startDashboardServer()
	},
}

func init() {
	rootCmd.AddCommand(dashCmd)
}

func startDashboardServer() {
	// Set up CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:8081"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Content-Type", "X-CSRF-Token"},
	})

	// Set up handlers
	http.HandleFunc("/", dashboardHandler)
	http.HandleFunc("/scan", k8s.HandleScanRequest)
	http.HandleFunc("/namespaces", k8s.HandleNamespaceListRequest)
	http.HandleFunc("/add-policy", k8s.HandleAddPolicyRequest)
	http.HandleFunc("/visualization", k8s.HandleVisualizationRequest)
	http.HandleFunc("/visualization/cluster", handleClusterVisualizationRequest)
	http.HandleFunc("/namespaces-with-policies", handleNamespacesWithPoliciesRequest)

	// Wrap the default serve mux with the CORS middleware
	handler := c.Handler(http.DefaultServeMux)

	// Start the server
	port := "8080"
	fmt.Printf("Starting dashboard server on http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Failed to start server: %v\n", err)
	}
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	// Serve the embedded frontend
	http.FileServer(statikFS).ServeHTTP(w, r)
}

// handleNamespacesWithPoliciesRequest handles the HTTP request for serving a list of namespaces with network policies.
func handleNamespacesWithPoliciesRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	namespaces, err := k8s.GatherNamespacesWithPolicies()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(struct {
		Namespaces []string `json:"namespaces"`
	}{Namespaces: namespaces}); err != nil {
		http.Error(w, "Failed to encode namespaces data", http.StatusInternalServerError)
	}
}

// handleClusterVisualizationRequest handles the HTTP request for serving cluster-wide visualization data.
func handleClusterVisualizationRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Call the function to gather cluster-wide visualization data
	clusterVizData, err := k8s.GatherClusterVisualizationData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(clusterVizData); err != nil {
		http.Error(w, "Failed to encode cluster visualization data", http.StatusInternalServerError)
	}
}
