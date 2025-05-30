curl -s -X POST http://localhost:8000/ocr \
  -H "Content-Type: application/json" \
  -d '{"device_id": "RF8RC00M8MF"}' \
  | jq 'map(del(.box))'