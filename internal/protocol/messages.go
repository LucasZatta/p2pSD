package protocol

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
)

// Tipos de mensagens do protocolo P2P
const (
	MsgTypeRequestBlock = "REQUEST_BLOCK"
	MsgTypeRequestInfo  = "REQUEST_INFO"
	MsgTypeBlockData    = "BLOCK_DATA"
	MsgTypePeerInfo     = "PEER_INFO"
	MsgTypeError        = "ERROR"
)

// Message é a interface base para todas as mensagens
type Message interface {
	GetType() string
}

// RequestBlockMsg - Cliente solicita um bloco específico
type RequestBlockMsg struct {
	Type    string `json:"type"`
	BlockID int    `json:"block_id"`
}

func (m *RequestBlockMsg) GetType() string {
	return m.Type
}

// RequestInfoMsg - Cliente solicita informações sobre blocos disponíveis
type RequestInfoMsg struct {
	Type string `json:"type"`
}

func (m *RequestInfoMsg) GetType() string {
	return m.Type
}

// BlockDataMsg - Servidor envia dados do bloco com checksum
type BlockDataMsg struct {
	Type     string `json:"type"`
	BlockID  int    `json:"block_id"`
	Data     []byte `json:"data"`
	Checksum string `json:"checksum"`
}

func (m *BlockDataMsg) GetType() string {
	return m.Type
}

// PeerInfoMsg - Servidor informa quais blocos possui
type PeerInfoMsg struct {
	Type            string `json:"type"`
	AvailableBlocks []int  `json:"available_blocks"`
	TotalBlocks     int    `json:"total_blocks"`
}

func (m *PeerInfoMsg) GetType() string {
	return m.Type
}

// ErrorMsg - Mensagem de erro
type ErrorMsg struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (m *ErrorMsg) GetType() string {
	return m.Type
}

// SendMessage envia uma mensagem via TCP
// Formato: [4 bytes tamanho][payload JSON]
func SendMessage(conn net.Conn, msg Message) error {
	// Serializa mensagem para JSON
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("erro ao serializar mensagem: %w", err)
	}

	// Envia tamanho (4 bytes, big-endian)
	size := uint32(len(data))
	if err := binary.Write(conn, binary.BigEndian, size); err != nil {
		return fmt.Errorf("erro ao enviar tamanho: %w", err)
	}

	// Envia payload
	if _, err := conn.Write(data); err != nil {
		return fmt.Errorf("erro ao enviar dados: %w", err)
	}

	return nil
}

// ReceiveMessage recebe uma mensagem via TCP
// Retorna um map[string]interface{} com os dados da mensagem
func ReceiveMessage(conn net.Conn) (map[string]interface{}, error) {
	// Lê tamanho (4 bytes)
	var size uint32
	if err := binary.Read(conn, binary.BigEndian, &size); err != nil {
		return nil, fmt.Errorf("erro ao ler tamanho: %w", err)
	}

	// Valida tamanho (máximo 16MB para segurança)
	if size > 16*1024*1024 {
		return nil, fmt.Errorf("mensagem muito grande: %d bytes", size)
	}

	// Lê payload
	data := make([]byte, size)
	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, fmt.Errorf("erro ao ler dados: %w", err)
	}

	// Deserializa JSON
	var msg map[string]interface{}
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("erro ao deserializar JSON: %w", err)
	}

	return msg, nil
}

// ParseMessage converte um map genérico para o tipo específico de mensagem
func ParseMessage(data map[string]interface{}) (Message, error) {
	msgType, ok := data["type"].(string)
	if !ok {
		return nil, fmt.Errorf("campo 'type' não encontrado ou inválido")
	}

	// Re-serializa e deserializa para o tipo correto
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	switch msgType {
	case MsgTypeRequestBlock:
		var msg RequestBlockMsg
		if err := json.Unmarshal(jsonData, &msg); err != nil {
			return nil, err
		}
		return &msg, nil

	case MsgTypeRequestInfo:
		var msg RequestInfoMsg
		if err := json.Unmarshal(jsonData, &msg); err != nil {
			return nil, err
		}
		return &msg, nil

	case MsgTypeBlockData:
		var msg BlockDataMsg
		if err := json.Unmarshal(jsonData, &msg); err != nil {
			return nil, err
		}
		return &msg, nil

	case MsgTypePeerInfo:
		var msg PeerInfoMsg
		if err := json.Unmarshal(jsonData, &msg); err != nil {
			return nil, err
		}
		return &msg, nil

	case MsgTypeError:
		var msg ErrorMsg
		if err := json.Unmarshal(jsonData, &msg); err != nil {
			return nil, err
		}
		return &msg, nil

	default:
		return nil, fmt.Errorf("tipo de mensagem desconhecido: %s", msgType)
	}
}

// NewRequestBlock cria uma mensagem de solicitação de bloco
func NewRequestBlock(blockID int) *RequestBlockMsg {
	return &RequestBlockMsg{
		Type:    MsgTypeRequestBlock,
		BlockID: blockID,
	}
}

// NewRequestInfo cria uma mensagem de solicitação de informações
func NewRequestInfo() *RequestInfoMsg {
	return &RequestInfoMsg{
		Type: MsgTypeRequestInfo,
	}
}

// NewBlockData cria uma mensagem com dados do bloco
func NewBlockData(blockID int, data []byte, checksum string) *BlockDataMsg {
	return &BlockDataMsg{
		Type:     MsgTypeBlockData,
		BlockID:  blockID,
		Data:     data,
		Checksum: checksum,
	}
}

// NewPeerInfo cria uma mensagem com informações do peer
func NewPeerInfo(availableBlocks []int, totalBlocks int) *PeerInfoMsg {
	return &PeerInfoMsg{
		Type:            MsgTypePeerInfo,
		AvailableBlocks: availableBlocks,
		TotalBlocks:     totalBlocks,
	}
}

// NewError cria uma mensagem de erro
func NewError(message string) *ErrorMsg {
	return &ErrorMsg{
		Type:    MsgTypeError,
		Message: message,
	}
}
