# Consul

Your guide to the Aritect ecosystem.

## System design

```
┌─────────────────────────────────────────────────────────────┐
│                        Telegram API                         │
└────────────────────────────┬────────────────────────────────┘
                             │
                             ▼
                    ┌────────────────┐
                    │     Client     │
                    │  (telebot.v3)  │
                    └────────┬───────┘
                             │
                             ▼
                    ┌────────────────┐
                    │     Router     │  ◄──── Config (env vars)
                    └────────┬───────┘
                             │
                ┌────────────┼────────────┐
                ▼            ▼            ▼
         ┌──────────┐  ┌──────────┐  ┌──────────┐
         │ Commands │  │ Commands │  │ Commands │
         │ /start   │  │ /website │  │ /chart   │
         │ /help    │  │ /ca      │  │ /arbiter │
         │ /id      │  │ /agartha │  │ ...      │
         └────┬─────┘  └────┬─────┘  └────┬─────┘
              │             │             │
              └─────────────┼─────────────┘
                            ▼
                   ┌─────────────────┐
                   │  Context Layer  │
                   │  - SendAnswer   │
                   │  - Logging      │
                   └────────┬────────┘
                            │
              ┌─────────────┼─────────────┐
              ▼             ▼             ▼
      ┌──────────┐   ┌────────────┐  ┌────────────┐
      │  Logger  │   │  Metrics   │  │    Store   │
      │          │   │(Prometheus)│  │  (LevelDB) │
      └──────────┘   └────────────┘  └────────────┘
                            │
                            ▼
                     ┌───────────────┐
                     │ :8080/metrics │
                     └───────────────┘
```

### Components

- **Bot Client**: Telegram API wrapper (gopkg.in/telebot.v3).
- **Router**: Message routing and command parsing.
- **Commands**: Business logic for each bot command.
- **Context**: Request context with helpers (SendAnswer, logging).
- **Config**: Environment-based configuration.
- **Store**: LevelDB for persistent data (recipients).
- **Metrics**: Prometheus metrics endpoint.
- **Logger**: Structured logging.

### Data flow

1. User sends message.
2. Bot Client receives update.
3. Router parses command and arguments.
4. Router executes corresponding command handler.
5. Command uses Context to send response.
6. Metrics and logs are recorded.

## License

The MIT License (MIT)

Copyright (c) 2025 Aritect

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
