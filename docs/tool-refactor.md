# Refatoração das ferramentas

Objetivos:

1. Criar uma tool por arquivo em internal/tools
2. Separar os testes de cada tum no arquivo de testes correspondente.
3. Rodar todos os testes a cada etapa para garantir que o funcionamento está correto, e fazer as correções apropriadas.

Melhorias da TUI para mostrar a execução das ferramentas:

1. A tool edit_file deve mostar um resumo das edições feitas: "-123 bytes, +456 bytes".
2. A tool go_test deve, em caso de erro, imprimir a saída para que possamos ver o que deu errado. Essa impressão deve ocorrer em texto esmaecido logo abaixo da saída padrão (curta).
3. A tool diff deve mostrar um resumo de estatísticas do que mudou.
4. A tool write_file deve mostrar o total de bytes escritos.

Melhoria na cobertura de testes:
1. Para as ferramentas Go Doc e Go Test devemos ter casos de testes apropriados.
