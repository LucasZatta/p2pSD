package peer

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/zatta/tp2-p2p/internal/checksum"
	"github.com/zatta/tp2-p2p/internal/metadata"
	"github.com/zatta/tp2-p2p/internal/protocol"
)

// NeighborInfo representa informações de um peer vizinho
type NeighborInfo struct {
	Address string // formato: "ip:port"
}

// Client representa o cliente que baixa blocos de outros peers
type Client struct {
	neighbors    []NeighborInfo
	blockManager *BlockManager
	metadata     *metadata.Metadata
	filePath     string
	logger       *log.Logger
	stopChan     chan struct{}
	wg           sync.WaitGroup
}

// NewClient cria um novo cliente
func NewClient(neighbors []NeighborInfo, blockManager *BlockManager, meta *metadata.Metadata, filePath string, logger *log.Logger) *Client {
	return &Client{
		neighbors:    neighbors,
		blockManager: blockManager,
		metadata:     meta,
		filePath:     filePath,
		logger:       logger,
		stopChan:     make(chan struct{}),
	}
}

// Start inicia o processo de download
func (c *Client) Start() {
	c.logger.Printf("[CLIENT] Iniciando download de %d vizinhos", len(c.neighbors))

	// Inicia uma goroutine para cada vizinho
	for _, neighbor := range c.neighbors {
		c.wg.Add(1)
		go c.downloadFromNeighbor(neighbor)
	}
}

// Wait aguarda todas as goroutines de download terminarem
func (c *Client) Wait() {
	c.wg.Wait()
	c.logger.Printf("[CLIENT] Todos os downloads concluídos")
}

// Stop para o cliente
func (c *Client) Stop() {
	close(c.stopChan)
}

// downloadFromNeighbor baixa blocos de um vizinho específico
func (c *Client) downloadFromNeighbor(neighbor NeighborInfo) {
	defer c.wg.Done()

	c.logger.Printf("[CLIENT] Conectando ao vizinho %s", neighbor.Address)

	// Tenta conectar com retry
	conn, err := c.connectWithRetry(neighbor.Address, 3, 2*time.Second)
	if err != nil {
		c.logger.Printf("[CLIENT] Falha ao conectar com %s: %v", neighbor.Address, err)
		return
	}
	defer conn.Close()

	c.logger.Printf("[CLIENT] Conectado a %s", neighbor.Address)

	// Loop de download até ter todos os blocos ou ser parado
	for {
		select {
		case <-c.stopChan:
			return
		default:
		}

		// Verifica se o download está completo
		if c.blockManager.IsDownloadComplete() {
			c.logger.Printf("[CLIENT] Download completo!")
			return
		}

		// Pega próximo bloco faltante
		blockID := c.blockManager.GetNextMissingBlock()
		if blockID == -1 {
			// Não há blocos faltantes
			return
		}

		// Tenta baixar o bloco
		if err := c.downloadBlock(conn, neighbor.Address, blockID); err != nil {
			c.logger.Printf("[CLIENT] Erro ao baixar bloco %d de %s: %v", blockID, neighbor.Address, err)

			// Se erro de conexão, tenta reconectar
			if isConnectionError(err) {
				c.logger.Printf("[CLIENT] Tentando reconectar com %s", neighbor.Address)
				conn.Close()
				conn, err = c.connectWithRetry(neighbor.Address, 3, 2*time.Second)
				if err != nil {
					c.logger.Printf("[CLIENT] Falha ao reconectar com %s: %v", neighbor.Address, err)
					return
				}
			}

			// Pequena pausa antes de tentar outro bloco
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Bloco baixado com sucesso
		c.logger.Printf("[CLIENT] Bloco %d baixado de %s - Progresso: %.1f%%",
			blockID, neighbor.Address, c.blockManager.GetProgress()*100)
	}
}

// downloadBlock baixa um bloco específico
func (c *Client) downloadBlock(conn net.Conn, neighborAddr string, blockID int) error {
	// Envia requisição do bloco
	request := protocol.NewRequestBlock(blockID)
	if err := protocol.SendMessage(conn, request); err != nil {
		return fmt.Errorf("erro ao enviar REQUEST_BLOCK: %w", err)
	}

	// Recebe resposta
	msgData, err := protocol.ReceiveMessage(conn)
	if err != nil {
		return fmt.Errorf("erro ao receber resposta: %w", err)
	}

	// Parse resposta
	msg, err := protocol.ParseMessage(msgData)
	if err != nil {
		return fmt.Errorf("erro ao parsear resposta: %w", err)
	}

	// Verifica tipo de mensagem
	switch m := msg.(type) {
	case *protocol.BlockDataMsg:
		// Valida checksum
		if !checksum.ValidateBlockChecksum(m.Data, m.Checksum) {
			return fmt.Errorf("checksum inválido para bloco %d", blockID)
		}

		// Valida com metadados
		expectedBlock, err := c.metadata.GetBlock(blockID)
		if err != nil {
			return fmt.Errorf("erro ao obter metadados: %w", err)
		}

		if m.Checksum != expectedBlock.Hash {
			return fmt.Errorf("checksum não corresponde aos metadados")
		}

		// Escreve bloco no arquivo
		if err := checksum.WriteBlockToFile(c.filePath, blockID, c.metadata.BlockSize, m.Data); err != nil {
			return fmt.Errorf("erro ao escrever bloco: %w", err)
		}

		// Marca bloco como disponível
		c.blockManager.MarkBlockAvailable(blockID)

		return nil

	case *protocol.ErrorMsg:
		return fmt.Errorf("erro do servidor: %s", m.Message)

	default:
		return fmt.Errorf("tipo de mensagem inesperado: %T", msg)
	}
}

// connectWithRetry tenta conectar com retry
func (c *Client) connectWithRetry(address string, maxRetries int, retryDelay time.Duration) (net.Conn, error) {
	var conn net.Conn
	var err error

	for i := 0; i < maxRetries; i++ {
		conn, err = net.DialTimeout("tcp", address, 5*time.Second)
		if err == nil {
			return conn, nil
		}

		if i < maxRetries-1 {
			c.logger.Printf("[CLIENT] Falha ao conectar com %s (tentativa %d/%d): %v. Tentando novamente...",
				address, i+1, maxRetries, err)
			time.Sleep(retryDelay)
		}
	}

	return nil, fmt.Errorf("falha após %d tentativas: %w", maxRetries, err)
}

// isConnectionError verifica se é um erro de conexão
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	// Verifica tipos comuns de erros de conexão
	_, isNetErr := err.(net.Error)
	return isNetErr
}
