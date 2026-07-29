# Processo de Desenvolvimento

Este diretório contém a documentação do processo de desenvolvimento do projeto `ze`. O objetivo central do projeto foi a criação de um agente de IA alimentado 100% por modelos locais.

## Evolução do Projeto

O desenvolvimento seguiu uma abordagem iterativa iniciada pela fase de "bootstrap", utilizando o documento "versão zero" para construir uma versão mínima do `ze` capaz de "auto-melhoramento". 

A partir desse ponto, o agente foi utilizado para implementar melhorias no próprio código, como a criação de novas ferramentas, adição de logs e suporte a exibição de Markdown. Esse ciclo de desenvolvimento - onde o agente é usado para construir e refinar a si mesmo - também serviu para definir e formatar os padrões de uso da ferramenta.

## Otimização para LLMs Locais

O design do `ze` foi cuidadosamente otimizado para o uso de LLMs locais, priorizando:
- **Simplicidade de Ferramentas:** Redução do número de ferramentas para evitar confusão do modelo.
- **Refinamento de Prompts:** Ajuste preciso das descrições para garantir o uso correto das capacidades.
- **Gestão de Contexto:** Prevenção de erros comuns, como a perda do histórico de raciocínio durante a interação com o modelo.

## Fluxo de Trabalho e Documentação

A documentação aqui presente registra cada fase dessa evolução. O fluxo de trabalho estabelecido permitiu que, após algumas iterações, o agente passasse a operar de forma quase autônoma para novas funcionalidades:

1. **Direcionamento:** O usuário aponta para uma especificação (ex: "ler a feature X em docs/03-features.md").
2. **Planejamento:** O agente traça um plano de desenvolvimento detalhado.
3. **Implementação:** O agente executa o plano passo a passo.

Este método permite a adição sistemática de novas funções mantendo a consistência do projeto.

## Timeline

- [01 - Refatoração de Ferramentas](01-refatoracao-de-ferramentas.md): Refatoração das ferramentas para garantir modularidade, testes individuais por ferramenta e melhorias na exibição de logs na TUI.
- [02 - Melhorias de UI e UX](02.0-melhorias-de-ui.md): Melhorias significativas na UI e UX, incluindo comandos multilinha, desativação de cores, aprimoramento do diff e suporte a entrada via stdin.
    - [02.1 - Refatoração do InputHandler](02.1-refatoracao.md): Plano para extrair a lógica de input para um componente testável, permitindo o uso de TDD para a feature de multilinhas.
    - [02.2 - Refatoração do Git Diff](02.2-diff.md): Plano para corrigir a exibição de estatísticas do `git diff`, tratando múltiplos arquivos, arquivos binários e contagem de arquivos não rastreados.
    - [02.3 - Modo Headless](02.3-headless.md): Plano para permitir o uso do `ze` em ambientes não interativos (pipes, scripts) através da detecção automática de TTY.
- [03 - Novas Features](03-features.md): Implementação de novas funcionalidades como execução de comandos shell, ferramenta WebFetch, renderização de Markdown, documentação completa do Go, persistência de sessão e logs de execução.
