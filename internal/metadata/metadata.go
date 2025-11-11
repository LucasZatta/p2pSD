package metadata

import (
	"encoding/json"
	"fmt"
	"os"
)

// BlockInfo representa informações sobre um bloco específico
type BlockInfo struct {
	ID     int    `json:"id"`
	Offset int64  `json:"offset"`
	Size   int    `json:"size"`
	Hash   string `json:"hash"`
}

// Metadata contém todas as informações sobre um arquivo compartilhado
type Metadata struct {
	FileName    string      `json:"file_name"`
	FileSize    int64       `json:"file_size"`
	BlockSize   int         `json:"block_size"`
	TotalBlocks int         `json:"total_blocks"`
	FileHash    string      `json:"file_hash"`
	Blocks      []BlockInfo `json:"blocks"`
}

// SaveToFile salva os metadados em um arquivo JSON
func (m *Metadata) SaveToFile(filePath string) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar metadados: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("erro ao escrever arquivo: %w", err)
	}

	return nil
}

// LoadFromFile carrega metadados de um arquivo JSON
func LoadFromFile(filePath string) (*Metadata, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo: %w", err)
	}

	var metadata Metadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("erro ao deserializar metadados: %w", err)
	}

	return &metadata, nil
}

// GetBlock retorna informações sobre um bloco específico
func (m *Metadata) GetBlock(blockID int) (*BlockInfo, error) {
	if blockID < 0 || blockID >= m.TotalBlocks {
		return nil, fmt.Errorf("ID de bloco inválido: %d (total: %d)", blockID, m.TotalBlocks)
	}

	return &m.Blocks[blockID], nil
}
