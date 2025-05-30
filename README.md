# Whiteout Survival Autopilot

**Autopilot for the “Whiteout Survival” mobile game**, providing end-to-end automation of gameplay via screen‐scraping, OCR, task scheduling and device control.

---

## Features

- **Automated gameplay** driven by screen captures + OCR (via PaddleOCR & Tesseract).  
- **Task scheduler** with TTL and priority per player profile.  
- **Bot-farm support**: coordinate multiple game instances/profiles on a single device.  
- **Redis-backed event bus** for inter-service messaging.  
- **Label Studio integration** for annotation / manual review workflows.  
- **Extensible rules engine** (CEL) and configuration via Viper.  

---

## Architecture

```text
┌──────────┐      ┌─────────┐      ┌───────────┐
│  Device  │◀────▶│  Redis  │◀────▶│ Go Autop. │
│ (ADB UI) │      │  (pub/sub)│     │  Service │
└──────────┘      └─────────┘      └───────────┘
       ▲                              │
       │                              ▼
       │                         ┌────────┐
       │                         │  OCR   │
       │                         │Service │
       │                         └────────┘
       │                              ▲
       │                              │
       └─────────────▶ LabelStudio ◀─┘
````

---

## Directory Structure

```
.
├── .adr-dir/            # Architecture Decision Records  
├── cmd/                 # Go entrypoints (autopilot binaries)  
├── internal/            # Go packages & business logic  
├── ocr/                 # Python OCR microservice (FastAPI)  
├── ops/                 # Operational scripts & configs  
├── references/          # Static assets (e.g. icons used for template matching)  
├── docs/                # Project documentation  
├── usecases/            # Defined gameplay use cases  
├── .gitignore          
├── docker-compose.yml   # Redis, Label Studio, OCR  
├── go.mod, go.sum       
└── note.md              # Roadmap & TODOs  
```

---

## How to Run

1. Launch **Whiteout Survival** on your device and ensure it is on the **main city screen**.
2. Install **ADB** and ensure your device is visible via `adb devices`.
3. Copy the device config example:

   ```bash
   cp db/devices.example.yaml db/devices.yaml
   ```
4. Edit `db/devices.yaml` and set the correct `device_id` (from `adb devices`).
5. Start the OCR server:

   ```bash
   cd ocr
   uv run screenshot_ocr_service.py --host 0.0.0.0 --port 8000
   ```
6. In a separate terminal, run the autopilot service:

   ```bash
   go run ./cmd/autopilot
   ```

---

## Documentation & Use Cases

* **docs/** — design docs, architecture diagrams, protocol specs.
* **usecases/** — step-by-step scenarios (e.g., “daily check-in”, “raid loop”).
* **.adr-dir/** — records of major architectural decisions.
