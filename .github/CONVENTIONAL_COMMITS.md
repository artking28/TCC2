# Diretrizes de Conventional Commits

Estas diretrizes padronizam mensagens de commit para facilitar entendimento, automação e histórico limpo.

## Formato

```
<tipo>(<escopo opcional>)<breaking_optional>: <assunto em uma linha>

<corpo opcional>

<rodapé opcional>
```

- `tipo`: uma das palavras reservadas listadas abaixo.
- `escopo` (opcional): área impactada (ex.: `cmd/indexador`, `utils`, `models`, `docs`, `build`).
- `breaking_optional`: use `!` após o tipo/escopo quando houver quebra de compatibilidade.
- `assunto`: frase curta, no imperativo, sem ponto final (até ~72 caracteres).
- `corpo`: o que e por quê da mudança. Quebre linhas em ~72 colunas.
- `rodapé`: referências a issues e notas (ex.: `Closes #123`, `BREAKING CHANGE:`).

## Tipos suportados

- `feat`: nova funcionalidade para o usuário.
- `fix`: correção de bug.
- `docs`: somente documentação.
- `style`: mudanças de formatação (sem lógica), ex.: espaços, vírgulas, lint.
- `refactor`: refatoração sem alterar comportamento externo.
- `perf`: melhorias de performance.
- `test`: adição/ajuste de testes.
- `build`: mudanças de build, ferramentas, dependências.
- `ci`: mudanças em pipelines e automações de CI.
- `chore`: tarefas diversas sem alterar código de produção.
- `revert`: reverte um commit anterior.

## Quebra de compatibilidade (BREAKING CHANGE)

Use `!` após o tipo/escopo e detalhe no corpo ou rodapé com o prefixo `BREAKING CHANGE:`.

Exemplos:
```
feat!: altera estrutura do índice para múltiplos shards

BREAKING CHANGE: reindexação completa é necessária; schema do banco mudou.
```
```
refactor(indexador)!: remove suporte a n-gramas com saltos

Motivo: simplificação do modelo e ganhos de performance no BM25.
BREAKING CHANGE: configurações antigas com `jumps` > 0 não são mais válidas.
```

## Exemplos

```
feat(indexador): adiciona suporte a BM25 no ranqueamento
```
```
fix(conversor): corrige extração de texto de PDFs scaneados
```
```
docs(tcc): adiciona template LaTeX e README do TCC
```
```
perf(utils): otimiza tokenização reduzindo alocações
```
```
chore(build): atualiza dependências do go.mod e go.sum
```
```
revert: revert "feat(indexador): cache em memória para consultas"
```

## Boas práticas

- Linguagem: mantenha consistência; assunto em tom imperativo (ex.: "adiciona", "corrige").
- Escopo: use nomes curtos e claros (ex.: `utils`, `models`, `cmd/indexador`, `tcc`).
- Assunto: evite detalhes do “como”; foque no “o que” e “por quê”.
- Corpo: explique contexto, decisões e impactos quando relevante.
- Referências: relacione issues/PRs no rodapé (`Closes #123`, `Refs #456`).

## Template de commit (opcional)

Você pode usar o arquivo `.github/commit-message-template` como modelo:

```
git config commit.template .github/commit-message-template
```

Assim, a mensagem será pré-preenchida no editor ao fazer `git commit`.

