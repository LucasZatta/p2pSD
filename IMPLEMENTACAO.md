# Implementação do Sistema P2P

## Visão Geral

Este projeto implementa um sistema de transferência de arquivos peer-to-peer (P2P) onde cada nó atua simultaneamente como cliente e servidor. O sistema foi desenvolvido em Go, aproveitando suas capacidades nativas de concorrência através de goroutines, o que permite gerenciar múltiplas conexões simultâneas de forma eficiente.

A arquitetura segue o modelo simétrico P2P, onde peers podem compartilhar blocos de arquivos entre si sem necessidade de um servidor central. Cada arquivo é fragmentado em blocos de tamanho configurável, e a integridade dos dados é garantida através de checksums SHA-256 tanto por bloco individual quanto para o arquivo completo.

## Arquitetura do Sistema

O sistema é organizado em quatro pacotes internos principais. O módulo de **protocolo** define as mensagens trocadas entre peers usando JSON sobre TCP, incluindo solicitações de blocos, informações de disponibilidade e transferência de dados. Para garantir que mensagens sejam corretamente delimitadas no stream TCP, cada mensagem é prefixada com seu tamanho em 4 bytes big-endian.

O módulo de **checksum** é responsável por toda validação de integridade. Ele calcula e verifica hashes SHA-256 tanto para blocos individuais quanto para o arquivo completo, garantindo que nenhuma corrupção de dados passe despercebida. A escrita de blocos no disco utiliza operações thread-safe com `WriteAt()`, permitindo que múltiplos blocos sejam escritos em paralelo sem conflitos.

O **gerenciador de metadados** mantém informações estruturadas sobre cada arquivo, incluindo seu tamanho total, tamanho de bloco, número de blocos e os checksums correspondentes. Esses metadados são armazenados em arquivos JSON que acompanham cada arquivo compartilhado.

No coração do sistema está o **gerenciador de blocos**, uma estrutura thread-safe que rastreia quais blocos já foram baixados e quais ainda faltam. Ele utiliza mutexes para coordenar o acesso concorrente e detecta automaticamente quando um download está completo.

Cada peer executa dois componentes simultaneamente. O **servidor TCP** aceita conexões de outros peers e responde a solicitações de informação sobre blocos disponíveis ou envia dados de blocos específicos. O **cliente TCP** conecta-se a peers vizinhos para baixar blocos faltantes, gerenciando automaticamente reconexões e retries em caso de falhas.

## Modos de Operação

Um peer pode operar em dois modos distintos. No modo **seeder**, o peer já possui o arquivo completo e apenas compartilha blocos com outros peers. No modo **leecher**, o peer inicia sem o arquivo e baixa blocos de seus vizinhos. Após completar o download e validar a integridade do arquivo, o leecher automaticamente se torna um seeder, compartilhando os blocos recém-baixados com outros peers.

## Executáveis

O projeto gera dois binários principais. O executável **peer** é a aplicação principal que pode ser configurada via arquivo JSON ou flags de linha de comando. Ele suporta logging configurável, tanto para arquivo quanto para stdout, e implementa graceful shutdown para encerrar conexões de forma limpa.

O executável **genfile** é uma ferramenta auxiliar que gera arquivos de teste com padrões reconhecíveis. Cada bloco gerado possui um cabeçalho identificador seguido de dados únicos baseados no ID do bloco, facilitando a validação e debugging. O genfile também cria automaticamente os arquivos de metadados correspondentes.

## Cenários de Teste

O sistema foi testado em dois cenários principais que avaliam diferentes aspectos da implementação.

### Cenário 1: 2 Peers, Blocos de 1KB

| Teste  | Arquivo | Tamanho | Blocos | Objetivo                     |
|--------|---------|---------|--------|------------------------------|
| Test A | File A  | 10KB    | 10     | Poucos blocos                |
| Test B | File B  | 1MB     | 1024   | Fragmentação razoável        |
| Test C | File C  | 10MB    | 10240  | Grandes transferências       |


### Cenário 2: 4 Peers, Blocos de 4KB

| Teste  | Arquivo | Tamanho | Blocos | Objetivo                     |
|--------|---------|---------|--------|------------------------------|
| Test A | File A  | 20KB    | 5      | Poucos blocos                |
| Test B | File B  | 5MB     | 1280   | Fragmentação razoável        |
| Test C | File C  | 20MB    | 5120   | Grandes transferências       |

Com quatro peers e blocos maiores de 4KB, este cenário testa a escalabilidade do sistema e o impacto do tamanho de bloco no throughput. A presença de três leechers simultâneos permite avaliar como o seeder gerencia contenção e múltiplas conexões paralelas.

A comparação entre cenários revela o impacto do número de peers e do tamanho de bloco na performance geral do sistema.

## Como Usar

O uso do sistema segue um fluxo simples. Primeiro, compile os binários com `go build` para gerar os executáveis `peer` e `genfile`. Em seguida, use o script `test/genfiles.sh` para gerar todos os arquivos de teste necessários com seus metadados.

Para executar um teste, navegue até o diretório do teste desejado (por exemplo, `test/scenarios/scenario1/test_a`) e execute `./run.sh` para iniciar os peers. Em outro terminal, você pode monitorar o progresso através dos logs gerados. Quando o download completar, execute `./verify.sh` para validar a integridade dos arquivos baixados. Por fim, use `./stop.sh` para encerrar todos os peers.

Cada script de verificação compara tanto o tamanho quanto o checksum SHA-256 dos arquivos baixados com os originais, garantindo que a transferência foi bem-sucedida e os dados estão íntegros.

## Estrutura do Projeto

```
tp2/
├── cmd/
│   ├── peer/              # Aplicação peer principal
│   └── genfile/           # Gerador de arquivos de teste
├── internal/
│   ├── protocol/          # Protocolo de comunicação TCP/JSON
│   ├── peer/              # Lógica do peer (cliente/servidor)
│   ├── metadata/          # Gerenciamento de metadados
│   └── checksum/          # Validação de integridade SHA-256
├── test/
│   ├── genfiles.sh        # Script para gerar arquivos
│   ├── files/             # Arquivos de teste gerados
│   └── scenarios/
│       ├── scenario1/     # 2 peers, blocos 1KB
│       │   ├── test_a/    # Arquivo 10KB
│       │   ├── test_b/    # Arquivo 1MB
│       │   └── test_c/    # Arquivo 10MB
│       └── scenario2/     # 4 peers, blocos 4KB
│           ├── test_a/    # Arquivo 20KB
│           ├── test_b/    # Arquivo 5MB
│           └── test_c/    # Arquivo 20MB
├── bin/                   # Executáveis compilados
└── configs/               # Exemplos de configuração
```

Cada diretório de teste contém as configurações dos peers em formato JSON, scripts para executar e parar os testes, e um script de verificação que valida a integridade dos arquivos transferidos.