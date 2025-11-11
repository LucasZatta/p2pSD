package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/zatta/tp2-p2p/internal/peer"
)

// Config representa a configuração do peer carregada do JSON
type Config struct {
	PeerID       string          `json:"peer_id"`
	ListenPort   int             `json:"listen_port"`
	Mode         string          `json:"mode"`
	FilePath     string          `json:"file_path"`
	MetadataPath string          `json:"metadata_path"`
	DownloadDir  string          `json:"download_dir"`
	Neighbors    []NeighborEntry `json:"neighbors"`
	LogFile      string          `json:"log_file,omitempty"`
}

// NeighborEntry representa um vizinho na configuração
type NeighborEntry struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

func main() {
	// Flags de linha de comando
	configFile := flag.String("config", "", "Arquivo de configuração JSON")
	peerID := flag.String("id", "", "ID do peer")
	port := flag.Int("port", 0, "Porta de escuta")
	mode := flag.String("mode", "", "Modo: seeder ou leecher")
	filePath := flag.String("file", "", "Caminho do arquivo")
	metadataPath := flag.String("metadata", "", "Caminho do arquivo de metadados")
	downloadDir := flag.String("download-dir", "./downloads", "Diretório de download")
	logFile := flag.String("log", "", "Arquivo de log (vazio = stdout)")
	flag.Parse()

	var config Config

	// Carrega configuração do arquivo JSON se fornecido
	if *configFile != "" {
		data, err := os.ReadFile(*configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao ler arquivo de configuração: %v\n", err)
			os.Exit(1)
		}

		if err := json.Unmarshal(data, &config); err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao parsear configuração JSON: %v\n", err)
			os.Exit(1)
		}
	}

	// Flags sobrescrevem configuração do arquivo
	if *peerID != "" {
		config.PeerID = *peerID
	}
	if *port != 0 {
		config.ListenPort = *port
	}
	if *mode != "" {
		config.Mode = *mode
	}
	if *filePath != "" {
		config.FilePath = *filePath
	}
	if *metadataPath != "" {
		config.MetadataPath = *metadataPath
	}
	if *downloadDir != "" {
		config.DownloadDir = *downloadDir
	}
	if *logFile != "" {
		config.LogFile = *logFile
	}

	// Valida configuração obrigatória
	if config.PeerID == "" {
		fmt.Fprintln(os.Stderr, "Erro: peer_id é obrigatório")
		flag.Usage()
		os.Exit(1)
	}
	if config.ListenPort == 0 {
		fmt.Fprintln(os.Stderr, "Erro: listen_port é obrigatório")
		flag.Usage()
		os.Exit(1)
	}
	if config.Mode == "" {
		fmt.Fprintln(os.Stderr, "Erro: mode é obrigatório (seeder ou leecher)")
		flag.Usage()
		os.Exit(1)
	}
	if config.MetadataPath == "" {
		fmt.Fprintln(os.Stderr, "Erro: metadata_path é obrigatório")
		flag.Usage()
		os.Exit(1)
	}

	// Configura logger
	var logger *log.Logger
	if config.LogFile != "" {
		// Log em arquivo
		logF, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao abrir arquivo de log: %v\n", err)
			os.Exit(1)
		}
		defer logF.Close()
		logger = log.New(logF, "", log.LstdFlags)
	} else {
		// Log em stdout
		logger = log.New(os.Stdout, "", log.LstdFlags)
	}

	logger.Printf("=== Peer P2P: %s ===", config.PeerID)
	logger.Printf("Modo: %s", config.Mode)
	logger.Printf("Porta: %d", config.ListenPort)

	// Converte modo
	var peerMode peer.PeerMode
	switch config.Mode {
	case "seeder":
		peerMode = peer.ModeSeeder
	case "leecher":
		peerMode = peer.ModeLeecher
	default:
		logger.Fatalf("Modo inválido: %s (use 'seeder' ou 'leecher')", config.Mode)
	}

	// Converte vizinhos
	neighbors := make([]peer.NeighborInfo, len(config.Neighbors))
	for i, n := range config.Neighbors {
		neighbors[i] = peer.NeighborInfo{
			Address: fmt.Sprintf("%s:%d", n.IP, n.Port),
		}
	}

	// Cria peer
	peerConfig := peer.PeerConfig{
		ID:           config.PeerID,
		Mode:         peerMode,
		Port:         config.ListenPort,
		FilePath:     config.FilePath,
		MetadataPath: config.MetadataPath,
		DownloadDir:  config.DownloadDir,
		Neighbors:    neighbors,
		Logger:       logger,
	}

	p, err := peer.NewPeer(peerConfig)
	if err != nil {
		logger.Fatalf("Erro ao criar peer: %v", err)
	}

	// Inicia peer
	if err := p.Start(); err != nil {
		logger.Fatalf("Erro ao iniciar peer: %v", err)
	}

	// Captura sinais de interrupção
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Se for leecher, aguarda download em goroutine
	if peerMode == peer.ModeLeecher {
		go func() {
			p.Wait()
			logger.Printf("[PEER] Download completo. Peer continua operando como seeder.")
			logger.Printf("[PEER] Pressione Ctrl+C para encerrar.")
		}()
	}

	// Aguarda sinal de interrupção
	<-sigChan
	logger.Println("\n[PEER] Recebido sinal de interrupção. Encerrando...")

	// Para peer
	p.Stop()

	logger.Printf("[PEER] Peer %s encerrado", config.PeerID)
}
