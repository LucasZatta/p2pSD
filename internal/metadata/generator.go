package metadata

import (
	"fmt"
	"path/filepath"

	"github.com/zatta/tp2-p2p/internal/checksum"
)

// GenerateFromFile gera metadados completos a partir de um arquivo
func GenerateFromFile(filePath string, blockSize int) (*Metadata, error) {
	// Obtém tamanho do arquivo
	fileSize, err := checksum.GetFileSize(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter tamanho do arquivo: %w", err)
	}

	// Calcula número total de blocos
	totalBlocks := checksum.CalculateTotalBlocks(fileSize, blockSize)

	// Calcula hash do arquivo completo
	fileHash, err := checksum.CalculateFileChecksum(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao calcular hash do arquivo: %w", err)
	}

	// Calcula checksums de todos os blocos
	blockChecksums, err := checksum.CalculateFileBlocksChecksums(filePath, blockSize)
	if err != nil {
		return nil, fmt.Errorf("erro ao calcular checksums dos blocos: %w", err)
	}

	// Cria lista de informações dos blocos
	blocks := make([]BlockInfo, totalBlocks)
	for i := 0; i < totalBlocks; i++ {
		offset := int64(i) * int64(blockSize)
		size := blockSize

		// Último bloco pode ser menor
		if i == totalBlocks-1 {
			remainingSize := int(fileSize - offset)
			if remainingSize < blockSize {
				size = remainingSize
			}
		}

		blocks[i] = BlockInfo{
			ID:     i,
			Offset: offset,
			Size:   size,
			Hash:   blockChecksums[i],
		}
	}

	// Cria estrutura de metadados
	metadata := &Metadata{
		FileName:    filepath.Base(filePath),
		FileSize:    fileSize,
		BlockSize:   blockSize,
		TotalBlocks: totalBlocks,
		FileHash:    fileHash,
		Blocks:      blocks,
	}

	return metadata, nil
}

// GenerateAndSave gera metadados e salva em arquivo
func GenerateAndSave(filePath string, blockSize int, outputPath string) error {
	metadata, err := GenerateFromFile(filePath, blockSize)
	if err != nil {
		return err
	}

	if err := metadata.SaveToFile(outputPath); err != nil {
		return err
	}

	return nil
}
