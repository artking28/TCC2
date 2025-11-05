# Trabalho de Conclusão de Curso (TCC)

Este diretório contém os arquivos relacionados à parte escrita do Trabalho de Conclusão de Curso.

## Estrutura

- `trabalho.tex` - Arquivo principal do trabalho em LaTeX, no padrão SENAC Santo Amaro
- `referencias.bib` - Arquivo com as referências bibliográficas (opcional, pode ser adicionado posteriormente)

## Sugestões de Evolução para o Código

Após análise do código atual, identifiquei algumas áreas que podem ser aprimoradas:

### 1. Arquitetura e Organização

1. **Padrão de Projeto Repository**: Considerar implementar o padrão Repository para melhorar a testabilidade e separação de responsabilidades.

2. **Inversão de Dependência**: Utilizar interfaces para desacoplar componentes e facilitar testes unitários.

3. **Configuração Centralizada**: Criar uma estrutura de configuração centralizada para parâmetros de sistema.

### 2. Testes e Qualidade de Código

1. **Testes Unitários**: Aumentar a cobertura de testes, especialmente para as funções de processamento de texto e algoritmos de ranqueamento.

2. **Testes de Integração**: Implementar testes de integração para os fluxos principais do sistema.

3. **Análise Estática**: Configurar ferramentas de análise estática (golangci-lint) para manter qualidade de código.

### 3. Performance e Otimização

1. **Cache de Resultados**: Implementar cache mais sofisticado para resultados de busca pré-computados.

2. **Otimização de Consultas SQL**: Rever consultas SQL para garantir índices adequados e desempenho.

3. **Concorrência**: Melhorar o uso de goroutines e canais para operações paralelas.

### 4. Funcionalidades Adicionais

1. **Interface Web**: Desenvolver uma interface web para facilitar a utilização do sistema.

2. **API REST**: Criar uma API REST para permitir integrações com outros sistemas.

3. **Exportação de Resultados**: Adicionar funcionalidades para exportar resultados em diferentes formatos (JSON, CSV).

### 5. Manutenibilidade

1. **Logging**: Implementar sistema de logging estruturado para melhor monitoramento e debug.

2. **Documentação**: Melhorar a documentação do código com comentários claros e exemplos de uso.

3. **Gerenciamento de Erros**: Padronizar o tratamento de erros com contexto adequado.

### 6. Segurança

1. **Validação de Entrada**: Implementar validação robusta para entradas do usuário.

2. **Sanitização**: Sanitizar dados antes de inserção em banco de dados para prevenir injeção SQL.

### 7. Monitoramento e Métricas

1. **Métricas de Desempenho**: Adicionar coleta de métricas para monitorar desempenho dos algoritmos.

2. **Avaliação de Resultados**: Implementar avaliações automáticas de precisão e revocação.

Essas melhorias podem ser implementadas iterativamente ao longo do desenvolvimento do TCC para elevar a qualidade do sistema e cumprir com os requisitos acadêmicos.