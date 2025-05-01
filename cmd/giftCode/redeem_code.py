#!/usr/bin/env python3
"""
Redeems a gift code for Whiteout Survival.

Новые возможности:
  • передавать игроков прямо в CLI:  --fid 424639700  --fid 401227964,420195101
  • если --fid указан хотя-бы раз — player.json НЕ читается

Старые опции (-c, -r, --restart) работают как прежде.
"""

import argparse, hashlib, json, sys, time, base64
from os.path import exists

import ddddocr, requests
from requests.adapters import HTTPAdapter, Retry

# ────────────────────────── CLI ──────────────────────────
parser = argparse.ArgumentParser()
parser.add_argument("-c", "--code", required=True)
parser.add_argument("-f", "--player-file", dest="player_file",
                    default="player.json", help=argparse.SUPPRESS)
parser.add_argument("-r", "--results-file", dest="results_file",
                    default="results.json")
parser.add_argument("--restart", action="store_true")
parser.add_argument("--fid", action="append",
                    help="player ID; можно несколько раз или через запятую")
args = parser.parse_args()

# ─────── формируем список players ───────
if args.fid:
    fids = []
    for token in args.fid:
        fids.extend(s.strip() for s in token.split(",") if s.strip())
    players = [{"id": fid, "original_name": fid} for fid in fids]
else:
    with open(args.player_file, encoding="utf-8") as fp:
        players = json.load(fp)

# ─────── результаты (как было) ───────
results = []
if exists(args.results_file):
    with open(args.results_file, encoding="utf-8") as fp:
        results = json.load(fp)

found = next((r for r in results if r["code"] == args.code), None)
if not found:
    print(f"New code: {args.code} — adding to results file and processing.")
    found = {"code": args.code, "status": {}}
    results.append(found)
result = found

# ─────── константы ───────
URL  = "https://wos-giftcode-api.centurygame.com/api"
SALT = "tB87#kPtkxqOS2"
HEAD = {"Content-Type": "application/x-www-form-urlencoded",
        "Accept": "application/json"}

# HTTP с повторами
session = requests.Session()
session.mount("https://", HTTPAdapter(
    max_retries=Retry(total=5, backoff_factor=1, status_forcelist=[429])))

ocr = ddddocr.DdddOcr(show_ad=False)

def solve(b64: str) -> str:
    if "," in b64:
        b64 = b64.split(",")[1]
    return ocr.classification(base64.b64decode(b64)).upper()

def md5(s: str) -> str:
    return hashlib.md5(s.encode()).hexdigest()

# ─────── счётчики ───────
cnt_ok = cnt_already = cnt_err = 0

# ─────── основной цикл ───────
for idx, p in enumerate(players, 1):
    print(f"\x1b[K{idx}/{len(players)} redeeming for {p['original_name']}",
          end="\r", flush=True)

    if result["status"].get(p["id"]) == "Successful" and not args.restart:
        cnt_already += 1
        continue

    ts = str(time.time_ns())
    login = {"fid": p["id"], "time": ts,
             "sign": md5(f"fid={p['id']}&time={ts}{SALT}")}

    if session.post(f"{URL}/player", data=login, headers=HEAD, timeout=30).json().get("msg") != "success":
        print(f"\nLogin failed for {p['original_name']} / {p['id']}")
        cnt_err += 1
        continue

    cap_req = {"fid": p["id"], "init": "0", "time": ts,
               "sign": md5(f"fid={p['id']}&init=0&time={ts}{SALT}")}
    cap_resp = session.post(f"{URL}/captcha", data=cap_req,
                            headers=HEAD, timeout=30).json()
    if cap_resp.get("msg") != "SUCCESS":
        print(f"\nCaptcha failed for {p['original_name']} / {p['id']}")
        cnt_err += 1
        continue

    captcha_code = solve(cap_resp["data"]["img"])
    ts2 = str(time.time_ns())
    redeem = {
        "fid": p["id"], "cdk": args.code,
        "captcha_code": captcha_code, "time": ts2
    }
    redeem["sign"] = md5(
        f"captcha_code={captcha_code}&cdk={args.code}&fid={p['id']}&time={ts2}{SALT}"
    )

    resp = session.post(f"{URL}/gift_code", data=redeem,
                        headers=HEAD, timeout=30).json()
    ec = resp.get("err_code")

    if ec == 40014:
        print("\nThe gift code doesn't exist!"); sys.exit(1)
    if ec == 40007:
        print("\nThe gift code is expired!");   sys.exit(1)

    if ec in (40008, 40011):              # already
        cnt_already += 1
        result["status"][p["id"]] = "Successful"
    elif ec == 20000:                     # success
        cnt_ok += 1
        result["status"][p["id"]] = "Successful"
    else:                                 # error
        cnt_err += 1
        result["status"][p["id"]] = "Unsuccessful"
        print(f"\nError: {resp}")

# ─────── save results ───────
with open(args.results_file, "w", encoding="utf-8") as fp:
    json.dump(results, fp)

print(f"\nSuccessfully claimed: {cnt_ok}  •  Already claimed: {cnt_already}  •  Errors: {cnt_err}")
