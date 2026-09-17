# Solution: Pipelined Multithreaded Lexer/Parser
- **Design Pattern:** Producer-Consumer architecture using Go channels for inter-stage communication.
- **Lexer (Producer):** Analyzes character-by-character to identify tokens. Supports multi-threaded tokenization for large source files.
- **Parser (Consumer):** Parses the token stream from the channel into the AST.
- **Performance Flag:** `-m <threads>` enables multi-threaded Lexer/Parser stages, offloading tokenization/parsing of disjoint file segments (or lines) to a thread pool.
- **Safety:** Each thread operates within its own Memory Region, ensuring no race conditions during the tokenization of the input buffer.
