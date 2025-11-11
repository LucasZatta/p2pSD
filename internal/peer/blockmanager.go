package peer

import (
	"sync"
)

// BlockManager gerencia o estado dos blocos de um arquivo de forma thread-safe
type BlockManager struct {
	totalBlocks      int
	availableBlocks  map[int]bool // map[blockID]isAvailable
	mu               sync.RWMutex
	downloadComplete bool
}

// NewBlockManager cria um novo gerenciador de blocos
func NewBlockManager(totalBlocks int) *BlockManager {
	return &BlockManager{
		totalBlocks:      totalBlocks,
		availableBlocks:  make(map[int]bool),
		downloadComplete: false,
	}
}

// MarkBlockAvailable marca um bloco como disponível
func (bm *BlockManager) MarkBlockAvailable(blockID int) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.availableBlocks[blockID] = true

	// Verifica se o download está completo
	if len(bm.availableBlocks) == bm.totalBlocks {
		bm.downloadComplete = true
	}
}

// IsBlockAvailable verifica se um bloco está disponível
func (bm *BlockManager) IsBlockAvailable(blockID int) bool {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	return bm.availableBlocks[blockID]
}

// GetAvailableBlocks retorna lista de IDs dos blocos disponíveis
func (bm *BlockManager) GetAvailableBlocks() []int {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	blocks := make([]int, 0, len(bm.availableBlocks))
	for blockID := range bm.availableBlocks {
		blocks = append(blocks, blockID)
	}

	return blocks
}

// GetAvailableBlocksCount retorna o número de blocos disponíveis
func (bm *BlockManager) GetAvailableBlocksCount() int {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	return len(bm.availableBlocks)
}

// GetMissingBlocksCount retorna o número de blocos faltantes
func (bm *BlockManager) GetMissingBlocksCount() int {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	return bm.totalBlocks - len(bm.availableBlocks)
}

// IsDownloadComplete verifica se todos os blocos foram baixados
func (bm *BlockManager) IsDownloadComplete() bool {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	return bm.downloadComplete
}

// GetTotalBlocks retorna o número total de blocos
func (bm *BlockManager) GetTotalBlocks() int {
	return bm.totalBlocks
}

// MarkAllBlocksAvailable marca todos os blocos como disponíveis (para seeders)
func (bm *BlockManager) MarkAllBlocksAvailable() {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	for i := 0; i < bm.totalBlocks; i++ {
		bm.availableBlocks[i] = true
	}
	bm.downloadComplete = true
}

// GetProgress retorna o progresso do download (0.0 a 1.0)
func (bm *BlockManager) GetProgress() float64 {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	if bm.totalBlocks == 0 {
		return 0.0
	}

	return float64(len(bm.availableBlocks)) / float64(bm.totalBlocks)
}

// GetNextMissingBlock retorna o próximo bloco faltante (menor ID), ou -1 se todos estão disponíveis
func (bm *BlockManager) GetNextMissingBlock() int {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	for i := 0; i < bm.totalBlocks; i++ {
		if !bm.availableBlocks[i] {
			return i
		}
	}

	return -1
}
