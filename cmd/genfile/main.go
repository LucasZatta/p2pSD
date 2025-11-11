package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/zatta/tp2-p2p/internal/metadata"
)

func main() {
	// Flags
	output := flag.String("output", "", "Caminho do arquivo de saída")
	size := flag.String("size", "", "Tamanho do arquivo (ex: 10KB, 1MB, 10MB)")
	blockSize := flag.Int("block-size", 1024, "Tamanho do bloco em bytes")
	metadataOutput := flag.String("metadata", "", "Caminho do arquivo de metadados (padrão: <output>.meta.json)")
	flag.Parse()

	// Valida argumentos
	if *output == "" {
		fmt.Fprintln(os.Stderr, "Erro: -output é obrigatório")
		flag.Usage()
		os.Exit(1)
	}

	if *size == "" {
		fmt.Fprintln(os.Stderr, "Erro: -size é obrigatório (ex: 10KB, 1MB, 10MB)")
		flag.Usage()
		os.Exit(1)
	}

	// Parse tamanho
	fileSize, err := parseSize(*size)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao parsear tamanho: %v\n", err)
		os.Exit(1)
	}

	// Define caminho do metadata se não fornecido
	metaPath := *metadataOutput
	if metaPath == "" {
		metaPath = *output + ".meta.json"
	}

	log.Printf("Gerando arquivo: %s", *output)
	log.Printf("Tamanho: %d bytes (%.2f MB)", fileSize, float64(fileSize)/1024/1024)
	log.Printf("Tamanho do bloco: %d bytes", *blockSize)

	// Gera arquivo com padrão reconhecível
	if err := generateFile(*output, fileSize, *blockSize); err != nil {
		log.Fatalf("Erro ao gerar arquivo: %v", err)
	}

	log.Printf("Arquivo gerado: %s", *output)

	// Gera metadados
	log.Printf("Gerando metadados...")
	if err := metadata.GenerateAndSave(*output, *blockSize, metaPath); err != nil {
		log.Fatalf("Erro ao gerar metadados: %v", err)
	}

	log.Printf("Metadados gerados: %s", metaPath)
	log.Printf("✓ Concluído!")
}

// generateFile gera um arquivo com padrão reconhecível
func generateFile(filePath string, size int64, blockSize int) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo: %w", err)
	}
	defer file.Close()

	// Gera conteúdo com padrão reconhecível
	// Formato: cada bloco tem um cabeçalho identificador seguido de dados
	// Isso permite validar a remontagem correta dos blocos

	totalBlocks := int(size) / blockSize
	if int(size)%blockSize != 0 {
		totalBlocks++
	}

	buffer := make([]byte, blockSize)
	bytesWritten := int64(0)

	for blockID := 0; blockID < totalBlocks; blockID++ {
		// Calcula tamanho deste bloco (último pode ser menor)
		currentBlockSize := blockSize
		remaining := size - bytesWritten
		if remaining < int64(blockSize) {
			currentBlockSize = int(remaining)
		}

		// Preenche bloco com padrão reconhecível
		fillBlock(buffer[:currentBlockSize], blockID, currentBlockSize)

		// Escreve bloco
		n, err := file.Write(buffer[:currentBlockSize])
		if err != nil {
			return fmt.Errorf("erro ao escrever bloco %d: %w", blockID, err)
		}

		bytesWritten += int64(n)
	}

	return nil
}

// fillBlock preenche um bloco com padrão reconhecível
func fillBlock(block []byte, blockID int, size int) {
	// Formato do bloco:
	// [16 bytes] Cabeçalho: "BLOCK:<blockID>    \n"
	// [restante] Padrão repetitivo baseado no blockID

	// Escreve cabeçalho
	header := fmt.Sprintf("BLOCK:%06d     \n", blockID)
	if len(block) < len(header) {
		// Bloco muito pequeno, apenas preenche com o que couber
		copy(block, header[:len(block)])
		return
	}
	copy(block, header)

	// Preenche restante com padrão baseado no blockID
	// Usa um padrão que se repete mas é único por bloco
	pattern := generatePattern(blockID)

	pos := len(header)
	for pos < size {
		n := copy(block[pos:], pattern)
		pos += n
	}
}

// generatePattern gera um padrão de bytes baseado no ID do bloco
func generatePattern(blockID int) []byte {
	// Cria um padrão de 64 bytes único para cada bloco
	pattern := make([]byte, 64)

	// Padrão: repetição do blockID e números sequenciais
	for i := 0; i < 64; i++ {
		pattern[i] = byte((blockID*64 + i) % 256)
	}

	return pattern
}

// parseSize converte string de tamanho para bytes
func parseSize(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(sizeStr)
	sizeStr = strings.ToUpper(sizeStr)

	multiplier := int64(1)
	value := sizeStr

	if strings.HasSuffix(sizeStr, "KB") {
		multiplier = 1024
		value = strings.TrimSuffix(sizeStr, "KB")
	} else if strings.HasSuffix(sizeStr, "MB") {
		multiplier = 1024 * 1024
		value = strings.TrimSuffix(sizeStr, "MB")
	} else if strings.HasSuffix(sizeStr, "GB") {
		multiplier = 1024 * 1024 * 1024
		value = strings.TrimSuffix(sizeStr, "GB")
	} else if strings.HasSuffix(sizeStr, "B") {
		multiplier = 1
		value = strings.TrimSuffix(sizeStr, "B")
	}

	var size int64
	_, err := fmt.Sscanf(value, "%d", &size)
	if err != nil {
		return 0, fmt.Errorf("formato inválido: %s", sizeStr)
	}

	return size * multiplier, nil
}
