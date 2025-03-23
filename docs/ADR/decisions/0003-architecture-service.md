# 3. architecture service

Date: 2025-03-23

## Status

Accepted

## Context

The issue motivating this decision, and any context that influences or constrains the decision.

## Decision

Schema events:

```csharp
Game Screenshot
       │
       ▼
[OCR + image parsing] → GameState ──┐
                                    ▼
                           [CEL condition eval]
                                    │
                                    ▼ true
                            [Scenario selected]
                                    │
                                    ▼
                          [FSM state check]
                                    │
                                    ▼ (если нужно)
                             [FSM transitions]
                                    │
                                    ▼
                            [Execute Scenario]
```

## Consequences

What becomes easier or more difficult to do and any risks introduced by the change that will need to be mitigated.
