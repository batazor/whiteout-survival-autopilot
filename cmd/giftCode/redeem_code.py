#!/usr/bin/env python3
"""
Redeem one Whiteout-Survival gift code for a single player.

Usage:
    uv run redeem_code.py -c HappyMay --fid 424639700
Stdout:
    SUCCESS | ALREADY_RECEIVED | CDK_NOT_FOUND | ERROR
"""

import argparse, hashlib, time, base64, sys
import ddddocr, requests
from requests.adapters import HTTPAdapter, Retry

# ── CLI ─────────────────────────────────────────────────────────
p = argparse.ArgumentParser()
p.add_argument("-c", "--code", required=True, help="gift code")
p.add_argument("--fid", required=True,      help="player ID")
args = p.parse_args()

FID  = args.fid
CODE = args.code

# ── const ───────────────────────────────────────────────────────
URL  = "https://wos-giftcode-api.centurygame.com/api"
SALT = "tB87#kPtkxqOS2"
HEAD = {"Content-Type": "application/x-www-form-urlencoded",
        "Accept": "application/json"}

sess = requests.Session()
sess.mount("https://", HTTPAdapter(
    max_retries=Retry(total=5, backoff_factor=1, status_forcelist=[429])))

ocr = ddddocr.DdddOcr(show_ad=False)

def md5(s: str) -> str: return hashlib.md5(s.encode()).hexdigest()
def solve(b64: str) -> str:
    if "," in b64: b64 = b64.split(",")[1]
    return ocr.classification(base64.b64decode(b64)).upper()

# ── 1. login ────────────────────────────────────────────────────
ts = str(time.time_ns())
login = {"fid": FID, "time": ts, "sign": md5(f"fid={FID}&time={ts}{SALT}")}

if sess.post(f"{URL}/player", data=login, headers=HEAD, timeout=30
             ).json().get("msg") != "success":
    print("ERROR"); sys.exit(0)

# ── 2. captcha ──────────────────────────────────────────────────
cap_req = {"fid": FID, "init": "0", "time": ts,
           "sign": md5(f"fid={FID}&init=0&time={ts}{SALT}")}
cap = sess.post(f"{URL}/captcha", data=cap_req, headers=HEAD, timeout=30).json()
if cap.get("msg") != "SUCCESS":
    print("ERROR"); sys.exit(0)

captcha = solve(cap["data"]["img"])

# ── 3. redeem ───────────────────────────────────────────────────
ts2 = str(time.time_ns())
redeem = {"fid": FID, "cdk": CODE, "captcha_code": captcha, "time": ts2}
redeem["sign"] = md5(f"captcha_code={captcha}&cdk={CODE}&fid={FID}&time={ts2}{SALT}")

res = sess.post(f"{URL}/gift_code", data=redeem, headers=HEAD, timeout=30).json()
ec  = res.get("err_code")

if ec == 20000:
    print("SUCCESS")
elif ec in (40008, 40011):
    print("ALREADY_RECEIVED")
elif ec == 40014:
    print("CDK_NOT_FOUND")
else:
    print("ERROR")
