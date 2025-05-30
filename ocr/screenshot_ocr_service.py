#!/usr/bin/env python3
import os
import subprocess
from fastapi import Body
from pydantic import BaseModel
import cv2
import struct
import numpy as np
import time
import asyncio
import paddle
from concurrent.futures import ThreadPoolExecutor
from fastapi import FastAPI, HTTPException
from paddleocr import PaddleOCR
from pydantic import BaseModel
from typing import List, Optional
import pytesseract
from PIL import Image
from collections import Counter
from starlette.middleware.base import BaseHTTPMiddleware
import json
from fastapi import Request
from typing import AsyncIterator

# --- Middleware for logging requests ---------------------------------------
class RequestLoggingMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next):
        path = request.url.path
        start_time = time.time()

        # Читаем тело запроса
        body_bytes = await request.body()
        try:
            body = json.loads(body_bytes.decode())
        except Exception:
            body = body_bytes.decode(errors="ignore") or "non-json"

        print(f"[REQ] {path} ⬇️ {body}")

        # Переотправим тело запроса, потому что .body() можно прочитать только один раз
        async def receive():
            return {"type": "http.request", "body": body_bytes}

        response = await call_next(Request(request.scope, receive))

        # Сохраняем тело ответа (если есть)
        response_body = b""
        async for chunk in response.body_iterator:
            response_body += chunk

        # Восстанавливаем асинхронный итератор из байтов
        async def new_body_iterator() -> AsyncIterator[bytes]:
            yield response_body

        response.body_iterator = new_body_iterator()

        try:
            response_data = json.loads(response_body.decode())

            if isinstance(response_data, list):
                for item in response_data:
                    if isinstance(item, dict):
                        if "box" in item:
                            item["box"] = "[...]"
                        if "avg_color" in item:
                            item["avg_color"] = colorize(item["avg_color"], item["avg_color"])
                        if "bg_color" in item:
                            item["bg_color"] = colorize(item["bg_color"], item["bg_color"])
        except Exception:
            response_data = response_body.decode(errors="ignore") or "non-json"

        process_time = time.time() - start_time
        print(f"[RES] {path} ⬆️")
        for item in response_data:
           if isinstance(item, dict):
               text = item.get("text", "")
               score = item.get("score", "")
               avg_color = item.get("avg_color", "")
               bg_color = item.get("bg_color", "")
               print(f"  - text: {text}")
               print(f"    score: {score}")
               print(f"    avg_color: {avg_color}")
               print(f"    bg_color: {bg_color}")
        print(f"({process_time:.3f}s)")

        return response

def colorize(text: str, color_name: str) -> str:
    ANSI_COLORS = {
        "black": "\033[30m",
        "red": "\033[31m",
        "green": "\033[32m",
        "yellow": "\033[33m",
        "blue": "\033[34m",
        "gray": "\033[90m",
        "white": "\033[97m",
        "reset": "\033[0m",
    }
    color = ANSI_COLORS.get(color_name, "")
    reset = ANSI_COLORS["reset"]
    return f"{color}{text}{reset}"

# --- Data models -------------------------------------------------------------
class Region(BaseModel):
    x0: int
    y0: int
    x1: int
    y1: int

class Zone(BaseModel):
    box: List[List[int]]
    text: str
    score: float
    avg_color: str
    bg_color: str

class OcrRequest(BaseModel):
    device_id: Optional[str] = None
    debug_name: Optional[str] = None
    regions: Optional[List[Region]] = None

class FindRequest(BaseModel):
    image_name: str
    device_id: Optional[str] = None
    threshold: Optional[float] = 0.8
    debug_name: Optional[str] = None
    regions: Optional[List[Region]] = None

class FindImageResponse(BaseModel):
    found: bool
    boxes: List[List[List[int]]]

class WaitRequest(BaseModel):
    stop_words: List[str]
    device_id: Optional[str] = None
    timeout: float            # в секундах
    interval: float           # в секундах
    debug_name: Optional[str] = None
    regions: Optional[List[Region]] = None

# --- Configuration ----------------------------------------------------------
DEBUG_MODE = os.path.exists("/DEBUG")
SCREENSHOT_TTL = 0.1  # seconds
CPU_THREADS = os.cpu_count() or 1
OCR_VERSION = os.getenv("OCR_VERSION", "PP-OCRv4")
ICON_DIR = os.path.abspath(os.path.join(
    os.path.dirname(__file__), "..", "references", "icons"
))

# --- Globals ----------------------------------------------------------------
app = FastAPI()
app.add_middleware(RequestLoggingMiddleware)
EXECUTOR = ThreadPoolExecutor(max_workers=CPU_THREADS)
_screenshot_cache = {"ts": 0.0, "img": None}
ocr = None
_gray_templates = {}
COLOR_MAP = {
    "black":  np.array([0, 0, 0]),
    "white":  np.array([255, 255, 255]),
    "red":    np.array([255, 0, 0]),
    "green":  np.array([89, 179, 97]),
    "blue":   np.array([72, 85, 119]),
    "yellow": np.array([255, 255, 0]),
    "gray":   np.array([128, 128, 128]),
}
COLOR_DIST_MATRIX = None

# --- Initialization --------------------------------------------------------
def init_services():
    global ocr, COLOR_DIST_MATRIX, _gray_templates
    paddle.disable_static()  # Add this line to enable dynamic graph mode
    paddle.set_device("cpu")
    ocr = PaddleOCR(
        ocr_version=OCR_VERSION,
        use_angle_cls=False,
        lang='en',
        use_gpu=False,
        det_limit_side_len=1024,
        cpu_threads=CPU_THREADS,
        ir_optim=True,
        layout=False,
        table=False,
        formula=False,
    )

    _gray_templates.clear()
    for fn in os.listdir(ICON_DIR):
        if fn.lower().endswith(".png"):
            name = os.path.splitext(fn)[0]
            img = cv2.imread(os.path.join(ICON_DIR, fn), cv2.IMREAD_GRAYSCALE)
            if img is not None:
                _gray_templates[name] = img

    palette = np.stack(list(COLOR_MAP.values()))
    COLOR_DIST_MATRIX = np.sqrt(
        np.sum((palette[:, None] - palette[None, :])**2, axis=2)
    )

@app.on_event("startup")
def on_startup():
    init_services()

# --- Screenshot cache -------------------------------------------------------
def get_screenshot(device_id: str = None) -> np.ndarray:
    """
    Захватываем «сырые» кадры RGBA без промежуточных файлов:
      • adb exec-out screencap (raw mode)
      • разбираем 12-байтовый заголовок (width, height, format)
      • получаем w*h*4 байт пикселей RGBA
      • конвертируем в BGR для OpenCV
    """
    # 1) Собираем команду
    cmd = ["adb"]
    if device_id:
        cmd += ["-s", device_id]
    cmd += ["exec-out", "screencap"]

    # 2) Запускаем и читаем все байты
    raw = subprocess.check_output(cmd)

    # 3) Парсим заголовок: <width:uint32><height:uint32><format:uint32>
    if len(raw) < 12:
        raise RuntimeError("Unexpected screencap output: too short for header")
    w, h, fmt = struct.unpack("<III", raw[:12])
    if fmt != 1:  # 1 == RGBA_8888
        raise RuntimeError(f"Unsupported format code: {fmt}")

    # 4) Достаём пиксельные данные
    expected = w * h * 4
    body = raw[12:]
    if len(body) < expected:
        raise RuntimeError(f"Screencap truncated: got {len(body)} of {expected} bytes")

    # 5) Формируем NumPy-массив RGBA
    img_rgba = np.frombuffer(body[:expected], dtype=np.uint8).reshape((h, w, 4))

    # 6) Переводим RGBA → BGR
    img_bgr = cv2.cvtColor(img_rgba, cv2.COLOR_RGBA2BGR)
    return img_bgr

# --- Color picker -----------------------------------------------------------
def pick_color_fast(pixels: np.ndarray) -> str:
    """
    Корректно классифицирует серые с оттенком как gray, только насыщенные пиксели считаются цветными.
    """
    bgr = np.uint8([[pixels[::-1]]])  # BGR для OpenCV
    hsv = cv2.cvtColor(bgr, cv2.COLOR_BGR2HSV)[0, 0]
    h, s, v = int(hsv[0]), int(hsv[1]), int(hsv[2])

    # Белый/чёрный по низкой насыщенности и яркости
    if s < 35 and v > 220:
        return "white"
    if s < 35 and v < 70:
        return "black"
    # ВСЁ, что менее насыщенное (даже с холодным оттенком) — gray
    if s < 60:
        return "gray"

    # Теперь цвета по hue
    if (h < 10 or h >= 170):
        return "red"
    if 10 <= h < 35:
        return "yellow"
    if 35 <= h < 85:
        return "green"
    if 85 <= h < 140:
        return "blue"

    return "gray"

def pick_color_for_ring(pixels: np.ndarray) -> str:
    # pixels shape: (N, 3), RGB
    colors = [pick_color_fast(p) for p in pixels]
    most_common = Counter(colors).most_common(1)
    return most_common[0][0] if most_common else "gray"

# --- Обработка региона и распознавание текста --------------
def process_roi(img: np.ndarray, region: Region) -> List[Zone]:
    """
    OCR по заданному региону + определение avg_color текста и bg_color фона.
    Вместо жёсткого порога белого используется бинаризация Otsu,
    чтобы корректно ловить и жёлтый текст.
    """
    zones: List[Zone] = []
    h, w = img.shape[:2]

    # Нормализуем и проверяем регион
    x0 = max(0, min(region.x0, w))
    x1 = max(0, min(region.x1, w))
    y0 = max(0, min(region.y0, h))
    y1 = max(0, min(region.y1, h))
    if x1 <= x0 or y1 <= y0:
        return zones

    # Выделяем ROI и guard
    roi_full = img[y0:y1, x0:x1]
    if roi_full.size == 0:
        return zones

    # Попытаемся вызвать OCR до 5 раз при ошибке allocator'а
    raw = []
    last_err = None
    for attempt in range(1, 6):
        try:
            ocr_results = ocr.ocr(roi_full, cls=False, det=True, rec=True)
            raw = ocr_results[0] if ocr_results else []
            break
        except RuntimeError as e:
            msg = str(e)
            if "No allocator found for the place" in msg:
                print(f"[WARN] Paddle OCR allocator error (attempt {attempt}/3), reinit services and retry...")
                last_err = e
                # Переинициализируем модель
                init_services()
                time.sleep(1)
                continue
            else:
                # какая-то другая RuntimeError — пробрасываем сразу
                raise
    else:
        # вышли по exhausted attempts
        raise RuntimeError(f"OCR failed after retries: {last_err}")

    if not raw:
        return zones

    names = list(COLOR_MAP.keys())
    palette = np.stack(list(COLOR_MAP.values()), axis=0).astype(np.int16)

    for box, (text, score) in raw:
        # Сдвигаем box в координаты полного img
        pts = np.array(box, dtype=int)
        pts[:,0] += x0
        pts[:,1] += y0

        # Вырезаем точную внутреннюю обрезку
        xs, ys = pts[:,0], pts[:,1]
        bx0, bx1 = np.clip([xs.min(), xs.max()], 0, w)
        by0, by1 = np.clip([ys.min(), ys.max()], 0, h)
        if bx1 <= bx0 or by1 <= by0:
            continue

        roi = img[by0:by1, bx0:bx1]
        if roi.size == 0:
            continue

        # === avg_color текста через Otsu ===
        gray_roi = cv2.cvtColor(roi, cv2.COLOR_BGR2GRAY)
        _, text_mask = cv2.threshold(
            gray_roi, 0, 255,
            cv2.THRESH_BINARY + cv2.THRESH_OTSU
        )
        mean_val = cv2.mean(roi, mask=text_mask)
        avg_rgb = (int(mean_val[2]), int(mean_val[1]), int(mean_val[0]))
        avg_color = pick_color_fast(np.array(avg_rgb))

        # === bg_color по кольцевой зоне ===
        mask = np.zeros((h, w), dtype=np.uint8)
        cv2.fillPoly(mask, [pts.reshape(-1,1,2)], 255)
        kernel = cv2.getStructuringElement(cv2.MORPH_ELLIPSE, (7,7))
        outer = cv2.dilate(mask, kernel)
        border = cv2.bitwise_and(outer, cv2.bitwise_not(mask))
        ys_b, xs_b = np.where(border > 0)
        if len(xs_b) == 0:
            bg_color = "gray"
        else:
            # BGR->RGB
            ring_pixels = img[ys_b, xs_b][:, ::-1]
            bg_color = pick_color_for_ring(ring_pixels)

        zones.append(Zone(
            box=pts.tolist(),
            text=text,
            score=float(score),
            avg_color=avg_color,
            bg_color=bg_color,
        ))

    return zones

def batch_ocr(img: np.ndarray, regions: List[Region]) -> List[Zone]:
    """
    Проходим по каждому region и вызываем process_roi(img, region):
    в нём происходит вся обрезка и OCR.
    """
    results: List[Zone] = []
    for r in regions:
        results.extend(process_roi(img, r))
    return results

def iou(a: np.ndarray, b: np.ndarray) -> float:
    # a, b — по 4 точки [[x0,y0],[x1,y0],[x1,y1],[x0,y1]]
    ax0, ay0 = a[0]
    ax1, ay1 = a[2]
    bx0, by0 = b[0]
    bx1, by1 = b[2]
    ix0, iy0 = max(ax0, bx0), max(ay0, by0)
    ix1, iy1 = min(ax1, bx1), min(ay1, by1)
    iw, ih = max(0, ix1-ix0), max(0, iy1-iy0)
    inter = iw*ih
    areaA = (ax1-ax0)*(ay1-ay0)
    areaB = (bx1-bx0)*(by1-by0)
    return inter / float(areaA+areaB-inter) if areaA+areaB-inter>0 else 0

def nms_boxes(all_boxes: List[List[List[int]]], thresh: float=0.5) -> List[List[List[int]]]:
    """
    all_boxes: список полигонов 4 точек [[ [x0,y0],... ], ...]
    thresh: IoU-порог
    """
    keep = []
    for box in all_boxes:
        should_keep = True
        for k in keep:
            if iou(np.array(box), np.array(k)) > thresh:
                should_keep = False
                break
        if should_keep:
            keep.append(box)
    return keep

# --- FastAPI endpoints ------------------------------------------------------
@app.post("/ocr", response_model=List[Zone])
async def ocr_endpoint(req: OcrRequest):
    """
    Одноразовый OCR по списку регионов, с автоповтором при ошибке
    'No allocator found for the place' до 3 раз.
    """
    max_attempts = 5
    for attempt in range(1, max_attempts + 1):
        start = time.time()
        loop = asyncio.get_running_loop()
        try:
            # 1) Снимаем скрин
            screen = await loop.run_in_executor(None, get_screenshot, req.device_id)

            # 2) Если нет регионов — анализируем весь экран
            if not req.regions:
                req.regions = [Region(x0=0, y0=0, x1=screen.shape[1], y1=screen.shape[0])]

            all_results: List[Zone] = []
            h, w = screen.shape[:2]

            # 3) Проходим по каждому региону
            for r in req.regions:
                # нормализуем границы
                x0 = max(0, min(r.x0, w))
                x1 = max(0, min(r.x1, w))
                y0 = max(0, min(r.y0, h))
                y1 = max(0, min(r.y1, h))
                if x1 <= x0 or y1 <= y0:
                    continue

                # вызываем OCR
                zones: List[Zone] = await loop.run_in_executor(
                    None, process_roi, screen, Region(x0=x0, y0=y0, x1=x1, y1=y1)
                )

                if zones:
                    all_results.extend(zones)
                else:
                    # фон всего региона
                    roi = screen[y0:y1, x0:x1]
                    mean_val = cv2.mean(roi)
                    avg_rgb = np.array([int(mean_val[2]), int(mean_val[1]), int(mean_val[0])])
                    bg_color = pick_color_fast(avg_rgb)
                    box = [[x0, y0], [x1, y0], [x1, y1], [x0, y1]]
                    all_results.append(Zone(
                        box=box,
                        text="",
                        score=1.0,
                        avg_color="",
                        bg_color=bg_color,
                    ))

            # DEBUG-снимок
            if DEBUG_MODE:
                debug_path = os.path.join(
                    "out",
                    req.debug_name or f"debug_{int(time.time())}.png"
                )
                await loop.run_in_executor(None, cv2.imwrite, debug_path, screen)

            duration = time.time() - start
            print(f"[OCR] /ocr done in {duration:.3f}s "
                  f"({len(all_results)} regions processed)")
            return all_results

        except RuntimeError as e:
            msg = str(e)
            if "No allocator found for the place" in msg and attempt < max_attempts:
                wait_time = 1 * attempt  # например, 1, 2, 3, 4 секунды...
                print(f"[WARN] Paddle OCR allocator error (attempt {attempt}/{max_attempts}), "
                      f"reinit services, wait {wait_time}s and retry…")
                init_services()
                await asyncio.sleep(wait_time)
                continue

            print(f"[ERROR] OCR endpoint failed after {attempt} attempts: {msg}")
            raise HTTPException(status_code=500, detail=f"OCR failed after {attempt} attempts: {msg}")
    # если все попытки исчерпаны, здесь мы уже пробросили ошибку выше


@app.post("/find_image", response_model=FindImageResponse)
async def find_image_endpoint(req: FindRequest):
    tpl = _gray_templates.get(req.image_name)
    if tpl is None:
        raise HTTPException(status_code=404, detail=f"Template {req.image_name} not found")

    loop = asyncio.get_running_loop()
    screen = await loop.run_in_executor(None, get_screenshot, req.device_id)
    gray = await loop.run_in_executor(None, cv2.cvtColor, screen, cv2.COLOR_BGR2GRAY)

    raw_boxes = []
    search_regions = req.regions or [Region(x0=0, y0=0, x1=gray.shape[1], y1=gray.shape[0])]

    for r in search_regions:
        roi = gray[r.y0:r.y1, r.x0:r.x1]
        res = cv2.matchTemplate(roi, tpl, cv2.TM_CCOEFF_NORMED)
        ys, xs = np.where(res >= req.threshold)
        for row, col in zip(ys, xs):
            x0, y0 = col + r.x0, row + r.y0
            x1, y1 = x0 + tpl.shape[1], y0 + tpl.shape[0]
            raw_boxes.append([[x0, y0], [x1, y0], [x1, y1], [x0, y1]])

    # отбираем неперекрывающиеся
    filtered = nms_boxes(raw_boxes, thresh=0.5)
    return FindImageResponse(found=bool(filtered), boxes=filtered)


@app.post("/wait_for_text", response_model=List[Zone])
async def wait_for_text_endpoint(req: WaitRequest = Body(...)):
    """
    Опрос /ocr до тех пор, пока в одной из зон не встретится любое стоп-слово (case-insensitive),
    либо не выйдет время timeout. Интервал между запросами — interval.
    """
    start = time.time()
    loop = asyncio.get_running_loop()

    while True:
        # 1) Захват экрана
        screen = await loop.run_in_executor(None, get_screenshot, req.device_id)

        # 2) OCR в нужных регионах или по всему экрану
        if req.regions:
            zones = await loop.run_in_executor(None, batch_ocr, screen, req.regions)
        else:
            full = Region(x0=0, y0=0, x1=screen.shape[1], y1=screen.shape[0])
            zones = await loop.run_in_executor(None, process_roi, screen, full)

        # 3) Ищем стоп-слова (case-insensitive)
        found = False
        for z in zones:
            txt = z.text.lower()
            for w in req.stop_words:
                if w.lower() in txt:
                    found = True
                    break
            if found:
                break

        # 4) Если нашли — сохраняем DEBUG-кадр и возвращаем все зоны
        if found:
            if DEBUG_MODE and req.debug_name:
                debug_path = os.path.join("out", f"{req.debug_name}.png")
                await loop.run_in_executor(None, cv2.imwrite, debug_path, screen)
            return zones

        # 5) Проверяем таймаут
        if time.time() - start >= req.timeout:
            return []  # ничего не нашли

        # 6) Ждём interval
        await asyncio.sleep(req.interval)

# --- Main -------------------------------------------------------------------
if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
