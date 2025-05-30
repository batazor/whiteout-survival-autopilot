# Примеры ADB-свайпов из пресетов (экран 1080x2400, 300px)

| Куда едет контент | Команда                                                                 |
|-------------------|-------------------------------------------------------------------------|
| **Вправо**        | `adb -s RF8RC00M8MF shell input touchscreen swipe 540 1200 240 1200 400` |
| **Влево**         | `adb -s RF8RC00M8MF shell input touchscreen swipe 540 1200 840 1200 400` |
| **Вверх**         | `adb -s RF8RC00M8MF shell input touchscreen swipe 540 1200 540 1500 400` |
| **Вниз**          | `adb -s RF8RC00M8MF shell input touchscreen swipe 540 1200 540 900 400`  |

**Формат:**  
`adb -s <device_id> shell input touchscreen swipe <x1> <y1> <x2> <y2> <duration_ms>`

- `<x1> <y1>` — стартовая точка
- `<x2> <y2>` — конечная точка
- `<duration_ms>` — длительность свайпа

> Для других направлений/смещений — меняй координаты аналогично.