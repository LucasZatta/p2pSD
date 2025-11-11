package peer

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/zatta/tp2-p2p/internal/checksum"
	"github.com/zatta/tp2-p2p/internal/metadata"
)

// PeerMode define o modo de operação do peer
type PeerMode string

const (
	ModeSeeder  PeerMode = "seeder"  // Peer que já possui o arquivo completo
	ModeLeecher PeerMode = "leecher" // Peer que está baixando o arquivo
)

// Peer representa um nó P2P que atua como cliente e servidor
type Peer struct {
	ID           string
	Mode         PeerMode
	Port         int
	FilePath     string
	DownloadDir  string
	Metadata     *metadata.Metadata
	Neighbors    []NeighborInfo
	BlockManager *BlockManager
	Server       *Server
	Client       *Client
	Logger       *log.Logger
	startTime    time.Time
}

// PeerConfig contém a configuração de um peer
type PeerConfig struct {
	ID           string
	Mode         PeerMode
	Port         int
	FilePath     string
	MetadataPath string
	DownloadDir  string
	Neighbors    []NeighborInfo
	Logger       *log.Logger
}

// NewPeer cria um novo peer
func NewPeer(config PeerConfig) (*Peer, error) {
	// Carrega metadados
	meta, err := metadata.LoadFromFile(config.MetadataPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar metadados: %w", err)
	}

	// Cria block manager
	blockManager := NewBlockManager(meta.TotalBlocks)

	var filePath string

	// Configura baseado no modo
	if config.Mode == ModeSeeder {
		// Seeder: verifica se arquivo existe
		if _, err := os.Stat(config.FilePath); err != nil {
			return nil, fmt.Errorf("arquivo não encontrado: %w", err)
		}
		filePath = config.FilePath

		// Marca todos os blocos como disponíveis
		blockManager.MarkAllBlocksAvailable()
		config.Logger.Printf("[PEER] Modo Seeder - Arquivo completo disponível: %s", filePath)

	} else {
		// Leecher: prepara arquivo de download
		if err := os.MkdirAll(config.DownloadDir, 0755); err != nil {
			return nil, fmt.Errorf("erro ao criar diretório de download: %w", err)
		}

		filePath = fmt.Sprintf("%s/%s", config.DownloadDir, meta.FileName)

		// Cria arquivo vazio com tamanho correto
		if err := createEmptyFile(filePath, meta.FileSize); err != nil {
			return nil, fmt.Errorf("erro ao criar arquivo de download: %w", err)
		}

		config.Logger.Printf("[PEER] Modo Leecher - Arquivo preparado: %s", filePath)
	}

	// Cria servidor
	server := NewServer(config.Port, blockManager, meta, filePath, config.Logger)

	// Cria cliente (apenas para leechers com vizinhos)
	var client *Client
	if config.Mode == ModeLeecher && len(config.Neighbors) > 0 {
		client = NewClient(config.Neighbors, blockManager, meta, filePath, config.Logger)
	}

	peer := &Peer{
		ID:           config.ID,
		Mode:         config.Mode,
		Port:         config.Port,
		FilePath:     filePath,
		DownloadDir:  config.DownloadDir,
		Metadata:     meta,
		Neighbors:    config.Neighbors,
		BlockManager: blockManager,
		Server:       server,
		Client:       client,
		Logger:       config.Logger,
	}

	return peer, nil
}

// Start inicia o peer (servidor e cliente se necessário)
func (p *Peer) Start() error {
	p.startTime = time.Now()
	p.Logger.Printf("[PEER] Iniciando peer %s na porta %d (modo: %s)", p.ID, p.Port, p.Mode)

	// Inicia servidor
	if err := p.Server.Start(); err != nil {
		return fmt.Errorf("erro ao iniciar servidor: %w", err)
	}

	// Se for leecher, inicia cliente para download
	if p.Mode == ModeLeecher && p.Client != nil {
		p.Logger.Printf("[PEER] Iniciando download de %d vizinhos", len(p.Neighbors))
		p.Client.Start()

		// Aguarda download em goroutine separada
		go p.waitForDownloadCompletion()
	}

	return nil
}

// Stop para o peer
func (p *Peer) Stop() {
	p.Logger.Printf("[PEER] Parando peer %s", p.ID)

	if p.Client != nil {
		p.Client.Stop()
	}

	if p.Server != nil {
		p.Server.Stop()
	}
}

// Wait aguarda o download ser concluído (apenas para leechers)
func (p *Peer) Wait() {
	if p.Mode == ModeLeecher && p.Client != nil {
		p.Client.Wait()
	}
}

// waitForDownloadCompletion aguarda download e valida arquivo
func (p *Peer) waitForDownloadCompletion() {
	// Aguarda conclusão
	p.Client.Wait()

	elapsed := time.Since(p.startTime)
	p.Logger.Printf("[PEER] Download concluído em %s", elapsed)

	// Valida integridade do arquivo
	p.Logger.Printf("[PEER] Validando integridade do arquivo...")
	if err := p.validateFile(); err != nil {
		p.Logger.Printf("[PEER] ERRO: Falha na validação: %v", err)
		return
	}

	p.Logger.Printf("[PEER] ✓ Arquivo validado com sucesso!")
	p.Logger.Printf("[PEER] Agora atuando como seeder")

	// Atualiza modo para seeder
	p.Mode = ModeSeeder

	// Imprime estatísticas
	p.printStats(elapsed)
}

// validateFile valida a integridade do arquivo baixado
func (p *Peer) validateFile() error {
	// Valida tamanho
	fileSize, err := checksum.GetFileSize(p.FilePath)
	if err != nil {
		return fmt.Errorf("erro ao obter tamanho do arquivo: %w", err)
	}

	if fileSize != p.Metadata.FileSize {
		return fmt.Errorf("tamanho incorreto: esperado %d, obtido %d", p.Metadata.FileSize, fileSize)
	}

	// Valida checksum do arquivo completo
	valid, err := checksum.ValidateFileChecksum(p.FilePath, p.Metadata.FileHash)
	if err != nil {
		return fmt.Errorf("erro ao validar checksum: %w", err)
	}

	if !valid {
		return fmt.Errorf("checksum do arquivo não corresponde")
	}

	return nil
}

// printStats imprime estatísticas do download
func (p *Peer) printStats(elapsed time.Duration) {
	totalBytes := p.Metadata.FileSize
	throughputMBps := float64(totalBytes) / elapsed.Seconds() / 1024 / 1024

	p.Logger.Printf("[PEER] ========== ESTATÍSTICAS ==========")
	p.Logger.Printf("[PEER] Arquivo: %s", p.Metadata.FileName)
	p.Logger.Printf("[PEER] Tamanho: %d bytes (%.2f MB)", totalBytes, float64(totalBytes)/1024/1024)
	p.Logger.Printf("[PEER] Blocos: %d (tamanho: %d bytes)", p.Metadata.TotalBlocks, p.Metadata.BlockSize)
	p.Logger.Printf("[PEER] Tempo: %s", elapsed)
	p.Logger.Printf("[PEER] Throughput: %.2f MB/s", throughputMBps)
	p.Logger.Printf("[PEER] Checksum: %s", p.Metadata.FileHash)
	p.Logger.Printf("[PEER] ====================================")
}

// GetProgress retorna o progresso do download (0.0 a 1.0)
func (p *Peer) GetProgress() float64 {
	return p.BlockManager.GetProgress()
}

// IsDownloadComplete verifica se o download está completo
func (p *Peer) IsDownloadComplete() bool {
	return p.BlockManager.IsDownloadComplete()
}

// GetStats retorna estatísticas do peer
func (p *Peer) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"peer_id":          p.ID,
		"mode":             string(p.Mode),
		"port":             p.Port,
		"total_blocks":     p.BlockManager.GetTotalBlocks(),
		"available_blocks": p.BlockManager.GetAvailableBlocksCount(),
		"missing_blocks":   p.BlockManager.GetMissingBlocksCount(),
		"progress":         p.GetProgress(),
		"complete":         p.IsDownloadComplete(),
	}
}

// createEmptyFile cria um arquivo vazio com tamanho específico
func createEmptyFile(filePath string, size int64) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Define tamanho do arquivo
	if err := file.Truncate(size); err != nil {
		return err
	}

	return nil
}
