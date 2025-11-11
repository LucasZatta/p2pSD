package checksum

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// CalculateBlockChecksum calcula o SHA-256 de um bloco de dados
func CalculateBlockChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(hash[:])
}

// CalculateFileChecksum calcula o SHA-256 de um arquivo completo
func CalculateFileChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("erro ao calcular hash: %w", err)
	}

	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
}

// ValidateBlockChecksum valida se o checksum de um bloco está correto
func ValidateBlockChecksum(data []byte, expectedChecksum string) bool {
	actualChecksum := CalculateBlockChecksum(data)
	return actualChecksum == expectedChecksum
}

// ValidateFileChecksum valida se o checksum de um arquivo está correto
func ValidateFileChecksum(filePath string, expectedChecksum string) (bool, error) {
	actualChecksum, err := CalculateFileChecksum(filePath)
	if err != nil {
		return false, err
	}
	return actualChecksum == expectedChecksum, nil
}

// CalculateFileBlocksChecksums calcula checksums de todos os blocos de um arquivo
func CalculateFileBlocksChecksums(filePath string, blockSize int) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer file.Close()

	var checksums []string
	buffer := make([]byte, blockSize)

	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("erro ao ler bloco: %w", err)
		}

		// Calcula checksum do bloco lido (pode ser menor que blockSize no último bloco)
		blockData := buffer[:n]
		checksum := CalculateBlockChecksum(blockData)
		checksums = append(checksums, checksum)
	}

	return checksums, nil
}

// ReadBlockFromFile lê um bloco específico de um arquivo
func ReadBlockFromFile(filePath string, blockID int, blockSize int) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer file.Close()

	// Calcula offset
	offset := int64(blockID) * int64(blockSize)

	// Move para posição do bloco
	if _, err := file.Seek(offset, 0); err != nil {
		return nil, fmt.Errorf("erro ao buscar posição: %w", err)
	}

	// Lê bloco
	buffer := make([]byte, blockSize)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("erro ao ler bloco: %w", err)
	}

	return buffer[:n], nil
}

// WriteBlockToFile escreve um bloco em uma posição específica do arquivo
func WriteBlockToFile(filePath string, blockID int, blockSize int, data []byte) error {
	// Abre arquivo para escrita (cria se não existir)
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer file.Close()

	// Calcula offset
	offset := int64(blockID) * int64(blockSize)

	// Escreve na posição específica
	if _, err := file.WriteAt(data, offset); err != nil {
		return fmt.Errorf("erro ao escrever bloco: %w", err)
	}

	return nil
}

// GetFileSize retorna o tamanho de um arquivo em bytes
func GetFileSize(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, fmt.Errorf("erro ao obter informações do arquivo: %w", err)
	}
	return info.Size(), nil
}

// CalculateTotalBlocks calcula o número total de blocos para um arquivo
func CalculateTotalBlocks(fileSize int64, blockSize int) int {
	totalBlocks := int(fileSize) / blockSize
	if int(fileSize)%blockSize != 0 {
		totalBlocks++
	}
	return totalBlocks
}
