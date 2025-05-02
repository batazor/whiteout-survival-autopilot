#!/usr/bin/env python3
"""
Redeem one Whiteout-Survival gift code for a single player.

stdout: SUCCESS | ALREADY_RECEIVED | CDK_NOT_FOUND | ERROR <reason>
"""

import argparse, hashlib, time, base64, sys, requests, ddddocr
from requests.adapters import HTTPAdapter, Retry

# ── CLI ─────────────────────────────────────────────────────────
p = argparse.ArgumentParser()
p.add_argument("-c", "--code", required=True)
p.add_argument("--fid", required=True)
args = p.parse_args()
FID, CODE = args.fid, args.code

# ── consts ──────────────────────────────────────────────────────
URL  = "https://wos-giftcode-api.centurygame.com/api"
SALT = "tB87#kPtkxqOS2"
HEAD = {"Content-Type": "application/x-www-form-urlencoded",
        "Accept": "application/json"}

sess = requests.Session()
sess.mount("https://", HTTPAdapter(max_retries=Retry(total=5,
                                                    backoff_factor=1,
                                                    status_forcelist=[429])))

ocr = ddddocr.DdddOcr(show_ad=False)
def md5(s: str) -> str: return hashlib.md5(s.encode()).hexdigest()
def solve(b64: str) -> str:
    if "," in b64: b64 = b64.split(",")[1]
    return ocr.classification(base64.b64decode(b64)).upper()

def die(reason: str):
    print(f"ERROR {reason}"); sys.exit(0)

# 1) login ───────────────────────────────────────────────────────
ts = str(time.time_ns())
login = {"fid": FID, "time": ts, "sign": md5(f"fid={FID}&time={ts}{SALT}")}
if sess.post(f"{URL}/player", data=login, headers=HEAD, timeout=30
             ).json().get("msg") != "success":
    die("LOGIN")

# 2-3) captcha + redeem  (до 3 попыток на случай 40101 / 40103) ──
maxTries = 3
for attempt in range(1, maxTries + 1):
    ts = str(time.time_ns())
    cap_req = {"fid": FID, "init": "0", "time": ts,
               "sign": md5(f"fid={FID}&init=0&time={ts}{SALT}")}
    cap = sess.post(f"{URL}/captcha", data=cap_req, headers=HEAD, timeout=30).json()
    if cap.get("msg") != "SUCCESS":
        die("CAPTCHA_REQUEST")

    captcha = solve(cap["data"]["img"])

    ts2 = str(time.time_ns())
    redeem = {"fid": FID, "cdk": CODE, "captcha_code": captcha, "time": ts2}
    redeem["sign"] = md5(f"captcha_code={captcha}&cdk={CODE}&fid={FID}&time={ts2}{SALT}")

    res = sess.post(f"{URL}/gift_code", data=redeem, headers=HEAD, timeout=30).json()
    ec, msg = res.get("err_code"), res.get("msg", "")

    if ec == 20000:
        print("SUCCESS");               sys.exit(0)
    if ec in (40008, 40011):
        print("ALREADY_RECEIVED");      sys.exit(0)
    if ec == 40014:
        print("CDK_NOT_FOUND");         sys.exit(0)
    if ec == 40007:
            print("CDK_EXPIRED");           sys.exit(0)
    if ec in (40101, 40103):            # CAPTCHA CHECK TOO FREQUENT / ERROR
        if attempt < maxTries:
            time.sleep(1)               # маленькая пауза перед новой капчей
            continue
    # любые другие или лимит попыток
    die(f"REDEEM ec={ec} msg='{msg}'")
