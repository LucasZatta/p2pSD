package peer

import (
	"fmt"
	"log"
	"net"

	"github.com/zatta/tp2-p2p/internal/checksum"
	"github.com/zatta/tp2-p2p/internal/metadata"
	"github.com/zatta/tp2-p2p/internal/protocol"
)

// Server representa o servidor TCP do peer
type Server struct {
	port         int
	listener     net.Listener
	blockManager *BlockManager
	metadata     *metadata.Metadata
	filePath     string
	logger       *log.Logger
	stopChan     chan struct{}
}

// NewServer cria um novo servidor
func NewServer(port int, blockManager *BlockManager, meta *metadata.Metadata, filePath string, logger *log.Logger) *Server {
	return &Server{
		port:         port,
		blockManager: blockManager,
		metadata:     meta,
		filePath:     filePath,
		logger:       logger,
		stopChan:     make(chan struct{}),
	}
}

// Start inicia o servidor TCP
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("erro ao iniciar servidor na porta %d: %w", s.port, err)
	}

	s.listener = listener
	s.logger.Printf("[SERVER] Escutando na porta %d", s.port)

	go s.acceptConnections()

	return nil
}

// Stop para o servidor
func (s *Server) Stop() {
	close(s.stopChan)
	if s.listener != nil {
		s.listener.Close()
	}
	s.logger.Printf("[SERVER] Servidor parado")
}

// acceptConnections aceita novas conexões em loop
func (s *Server) acceptConnections() {
	for {
		select {
		case <-s.stopChan:
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				select {
				case <-s.stopChan:
					return
				default:
					s.logger.Printf("[SERVER] Erro ao aceitar conexão: %v", err)
					continue
				}
			}

			// Processa conexão em goroutine separada
			go s.handleConnection(conn)
		}
	}
}

// handleConnection trata uma conexão de cliente
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	s.logger.Printf("[SERVER] Nova conexão de %s", remoteAddr)

	// Loop para receber múltiplas requisições na mesma conexão
	for {
		// Recebe mensagem
		msgData, err := protocol.ReceiveMessage(conn)
		if err != nil {
			// Conexão fechada ou erro
			s.logger.Printf("[SERVER] Conexão fechada por %s", remoteAddr)
			return
		}

		// Parse mensagem
		msg, err := protocol.ParseMessage(msgData)
		if err != nil {
			s.logger.Printf("[SERVER] Erro ao parsear mensagem de %s: %v", remoteAddr, err)
			errMsg := protocol.NewError(fmt.Sprintf("Erro ao parsear mensagem: %v", err))
			protocol.SendMessage(conn, errMsg)
			continue
		}

		// Processa baseado no tipo
		switch m := msg.(type) {
		case *protocol.RequestInfoMsg:
			s.handleRequestInfo(conn, remoteAddr)

		case *protocol.RequestBlockMsg:
			s.handleRequestBlock(conn, remoteAddr, m.BlockID)

		default:
			s.logger.Printf("[SERVER] Tipo de mensagem desconhecido de %s", remoteAddr)
			errMsg := protocol.NewError("Tipo de mensagem não suportado")
			protocol.SendMessage(conn, errMsg)
		}
	}
}

// handleRequestInfo responde com informações sobre blocos disponíveis
func (s *Server) handleRequestInfo(conn net.Conn, remoteAddr string) {
	availableBlocks := s.blockManager.GetAvailableBlocks()
	totalBlocks := s.blockManager.GetTotalBlocks()

	s.logger.Printf("[SERVER] REQUEST_INFO de %s - Disponíveis: %d/%d", remoteAddr, len(availableBlocks), totalBlocks)

	response := protocol.NewPeerInfo(availableBlocks, totalBlocks)
	if err := protocol.SendMessage(conn, response); err != nil {
		s.logger.Printf("[SERVER] Erro ao enviar PEER_INFO para %s: %v", remoteAddr, err)
	}
}

// handleRequestBlock responde com dados do bloco solicitado
func (s *Server) handleRequestBlock(conn net.Conn, remoteAddr string, blockID int) {
	// Verifica se o bloco está disponível
	if !s.blockManager.IsBlockAvailable(blockID) {
		s.logger.Printf("[SERVER] REQUEST_BLOCK %d de %s - Bloco não disponível", blockID, remoteAddr)
		errMsg := protocol.NewError(fmt.Sprintf("Bloco %d não disponível", blockID))
		protocol.SendMessage(conn, errMsg)
		return
	}

	// Lê bloco do arquivo
	blockData, err := checksum.ReadBlockFromFile(s.filePath, blockID, s.metadata.BlockSize)
	if err != nil {
		s.logger.Printf("[SERVER] Erro ao ler bloco %d: %v", blockID, err)
		errMsg := protocol.NewError(fmt.Sprintf("Erro ao ler bloco: %v", err))
		protocol.SendMessage(conn, errMsg)
		return
	}

	// Calcula checksum do bloco
	blockChecksum := checksum.CalculateBlockChecksum(blockData)

	// Valida com checksum dos metadados
	expectedBlock, err := s.metadata.GetBlock(blockID)
	if err != nil {
		s.logger.Printf("[SERVER] Erro ao obter metadados do bloco %d: %v", blockID, err)
		errMsg := protocol.NewError("Erro ao obter metadados do bloco")
		protocol.SendMessage(conn, errMsg)
		return
	}

	if blockChecksum != expectedBlock.Hash {
		s.logger.Printf("[SERVER] AVISO: Checksum do bloco %d não corresponde aos metadados", blockID)
	}

	// Envia bloco
	s.logger.Printf("[SERVER] Enviando bloco %d (%d bytes) para %s", blockID, len(blockData), remoteAddr)
	response := protocol.NewBlockData(blockID, blockData, blockChecksum)
	if err := protocol.SendMessage(conn, response); err != nil {
		s.logger.Printf("[SERVER] Erro ao enviar BLOCK_DATA para %s: %v", remoteAddr, err)
	}
}
