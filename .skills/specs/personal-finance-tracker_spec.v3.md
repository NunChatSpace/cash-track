# personal-finance-tracker/spec.v2.md

## 1. Goal

ลด friction ในการจดการเงินส่วนตัว โดย:

1. ใช้ **แชทบอท** เป็นอินเตอร์เฟซหลัก  
   - พิมพ์ข้อความธรรมดาเกี่ยวกับรายจ่าย/รายรับ  
   - อัปโหลดรูปสลิป (bank / e-wallet / marketplace) แนบในแชทได้
2. มี **dashboard ดู metric** ว่าเงิน 100% ถูกใช้/รับมาจากไหนบ้าง (ตาม category, channel, etc.)
3. เผื่ออนาคตสามารถต่อยอดเป็น **planning / budgeting** ได้ โดยไม่ต้องรื้อ data model เดิม

หลักคิด:
- ใช้ LLM แค่ช่วย **แปลงข้อความ/สลิป → โครงสร้างข้อมูล (JSON)**  
- การคำนวณ metric / สรุปยอด / planning ทำแบบ **deterministic** ที่ backend และ DB
- ระบบใช้คนเดียว รันบน PC ตัวเองได้ (SQLite + script / service เบา ๆ)

---

## 2. Scope

### In-scope (MVP)

- แชทบอทส่วนตัว (single user) ที่รองรับ:
- เมื่อ LLM extract ได้ไม่ครบหรือไม่มั่นใจ: ให้เซฟเป็น transaction ที่มี `status = 'pending'` ทันที แล้วแสดงเป็น “transaction card” ให้เราเข้ามาแก้/confirm ภายหลัง (ไม่ถือว่าหายไปเฉย ๆ)

  - ข้อความธรรมดา เช่น  
    - "วันนี้กินข้าวมื้อแรกไป 55 บาท"  
    - "เมื่อวาน Shopee 320 บาท scb เครดิต"
  - ข้อความ + รูปสลิป (OCR text)
- ใช้ LLM (ผ่าน Codex CLI หรือเทียบเท่า) เพื่อ:
  - วิเคราะห์ intent:
    - `add_transaction` (เพิ่มรายการ)
    - `bill_payment` (จากสลิปจ่ายบิล)
    - `query_summary` (ถามสรุป)
    - `unknown`
  - extract field ต่าง ๆ เป็น JSON
- Backend:
  - validate JSON จาก LLM
  - บันทึก `transactions` ลง SQLite
  - คำนวณ summary / metric ตามช่วงเวลา
- Dashboard (เบื้องต้น):
  - breakdown รายจ่าย/รายรับตาม category (% ของ total)
  - breakdown ตาม channel (เช่น cash / scb / kbank / tmw)
  - filter ตามช่วงเวลา (เช่น เดือนนี้ / เดือนที่แล้ว / custom range)

### Out-of-scope (MVP)

- ไม่มีการดึงข้อมูลตรงจาก API ธนาคาร / e-wallet
- ไม่มี multi-user / share account
- ไม่มี planning / budgeting engine (เก็บข้อมูลให้พร้อมต่อยอดเฉย ๆ)
- ไม่การันตีว่า categorization ถูก 100% → user ยังต้องแก้ได้

### Future Scope (v2+)

- ตั้ง budget ต่อ category / ต่อเดือน
- แนะนำการใช้เงิน (เช่น overspending alert)
- rule-based auto-categorization (ต่อยอดจาก manual + LLM)

---

## 3. User Flow

### 3.1 บันทึกรายการผ่านข้อความ (Text-only)

1. User เปิดแชท (CLI / small UI)
2. พิมพ์ข้อความ เช่น  
   `วันนี้กินข้าวมื้อแรกไป 55 บาท เงินสด`
3. Backend ส่งข้อความไปให้ LLM adapter (Codex CLI)
4. LLM ส่งกลับ JSON (intent + transaction fields)
5. Backend:
   - validate field:
     - amount > 0
     - txn_date ไม่อยู่ในอนาคต (ถ้าไม่แน่ใจ → null)
   - ถ้า field ครบพอใช้ → บันทึกลง DB ทันที (status = `confirmed`)
   - ถ้า field ขาด/งง → ตอบ user ให้เติมข้อมูล หรือยืนยัน/แก้ไข ผ่านแชท (MVP อาจตอบว่า "ไม่เข้าใจ" ให้พิมพ์ใหม่)
6. ระบบตอบในแชท:
   - สรุปรายการที่บันทึกแล้ว (เช่น "บันทึกค่าอาหาร 55 บาท (cash) วันที่ 2026-01-31 แล้ว")

> หมายเหตุ: เพื่อให้ระบบเล็กพอ รันบน PC เดียว ใน MVP สามารถข้ามขั้นตอน UI ฟอร์มยืนยันแยกจอ แล้วใช้แชทเป็นที่ confirm/แก้ไขไปก่อน

---

### 3.2 บันทึกรายการผ่านรูปสลิป (OCR)

1. User อัปโหลดรูปสลิปในแชท + (optional) พิมพ์ note สั้น ๆ
2. Backend:
   - เก็บไฟล์ลง local disk (เช่น `./data/slips/...`)
   - รัน OCR (เช่น engine local หรือส่งออกไปก่อนก็ได้ แต่ใน spec นี้มองเป็น text-ready)
3. Backend เรียก LLM พร้อม OCR text + note
4. LLM ตอบ JSON:

   - `intent = "bill_payment"` หรือ `"add_transaction"`  
   - field ที่เกี่ยวข้อง: `txn_date`, `amount`, `channel`, `category`, `note`, `raw_text`, etc.

5. Backend validate:
   - ถ้าแน่ใจ (field ครบ ชัด) → save เป็น `confirmed`
   - ถ้าไม่แน่ใจ → save เป็น `pending` และบอก user ให้พิมพ์คำสั่งแก้ เช่น  
     `"แก้หมวดเป็น food"` หรือ `"เปลี่ยน amount เป็น 97"`

6. ระบบตอบในแชท:
   - สรุปสิ่งที่มันเข้าใจจากสลิป เช่น  
     "จับได้ว่าคุณจ่าย 96.00 THB ผ่าน TrueMoney ให้ร้านเคิมฉีหม่าล่า เมื่อ 2026-01-30 — บันทึกเป็นหมวดอาหารแล้ว"

---

### 3.3 ดู metric / summary ผ่านแชท

1. User ถามในแชท เช่น:
   - "เดือนนี้ใช้เงินไปเท่าไหร่"
   - "เดือนที่แล้วค่าอาหารเท่าไหร่"
   - "ปีนี้ Shopee หมดไปกี่บาท"
2. Backend ส่งข้อความไปให้ LLM เพื่อ:
   - แปลงคำถาม → intent `query_summary`
   - ระบุช่วงเวลา / filter / category ที่ต้องใช้
3. Backend:
   - รัน SQL กับ DB:
     - รวมยอดตาม filter ที่ LLM แปลมา
     - คำนวณสัดส่วน (%) ต่อ total
   - สร้างคำตอบในรูป text + data summary
4. ระบบตอบในแชท:
   - Text ธรรมดา เช่น  
     "เดือนนี้ (2026-01) คุณใช้ไป 12,340 บาท ในหมวดอาหาร 4,200 บาท (34%)"
   - (Optional) ถ้าต่อกับ dashboard UI ด้วย → front-end ไปเรียก API summary แสดงกราฟเอง

---

### 3.4 Dashboard (ดู metrics แบบรวม ๆ)

*ไม่จำเป็นต้องใช้ LLM*

1. User เปิดหน้าต่าง dashboard (HTML/desktop UI)
2. Backend มี endpoint เช่น:

   - `GET /api/dashboard/summary?from=2026-01-01&to=2026-01-31`
   - `GET /api/dashboard/by-category?from=...&to=...`
   - `GET /api/dashboard/by-channel?from=...&to=...`

3. Backend:
   - query DB
   - return JSON แบบพร้อมใช้วาด chart (labels + values + percentage)
4. Frontend แสดง:
   - Pie chart: breakdown ตาม category
   - Bar chart: breakdown ตาม channel
   - Time series: ยอดใช้ต่อวัน/สัปดาห์/เดือน (optional)

---

## 4. Data Model

### 4.1 Table: transactions

```sql
CREATE TABLE transactions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    txn_date        TEXT            NOT NULL,            -- 'YYYY-MM-DD'
    amount          REAL            NOT NULL,
    currency        TEXT            NOT NULL DEFAULT 'THB',
    direction       TEXT            NOT NULL,            -- 'income' | 'expense' | 'transfer'
    channel         TEXT            NOT NULL,            -- 'cash' | 'scb' | 'kbank' | 'tmw' | ...
    account_label   TEXT,                               -- label สั้นๆ เช่น 'scb-main', 'kbank-cc'
    category        TEXT,                               -- 'food' | 'rent' | 'shopping' | ...
    description     TEXT,                               -- สรุปสั้น ๆ
    slip_image_path TEXT,                               -- path ไฟล์ใน local (nullable สำหรับ text-only)
    raw_ocr_text    TEXT,                               -- full text จาก OCR (ถ้ามี)
    llm_confidence  REAL,                               -- 0.0 - 1.0 หรือ null ถ้าไม่ได้ใช้
    status          TEXT            NOT NULL DEFAULT 'confirmed', -- 'pending' | 'confirmed' | 'rejected'
    created_at      TEXT            NOT NULL,           -- ISO timestamp
    updated_at      TEXT            NOT NULL            -- ISO timestamp
);
```

หมายเหตุ:
- ใช้ `TEXT` แทน `DATE/TIMESTAMP` เพื่อให้ query ผ่าน SQLite ง่ายบน PC เดียว
- `direction` = ทำให้รองรับ income/expense ได้ตั้งแต่แรก → metric “100% ของเราใช้ไปกับอะไร” จะอ่านง่าย
- `account_label` = ให้แมปบัญชีเวลาเปลี่ยนธนาคาร/หมายเลขบัญชีในอนาคตได้ โดยไม่ต้องแก้ข้อมูลเก่า

### 4.2 (Optional) Table: categories

ถ้าต้องการ normalize category:

```sql
CREATE TABLE categories (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    code        TEXT UNIQUE NOT NULL,   -- เช่น 'food', 'rent', 'shopping'
    name        TEXT NOT NULL           -- display name
);
```

MVP สามารถ hardcode category ในโค้ดไปก่อนก็ได้ ไม่จำเป็นต้องมี table แยก

---

## 5. Minimal API Surface (สำหรับต่อ UI / script ง่าย ๆ)

แม้ระบบจะรันบน PC ตัวเดียว แต่การนิยาม API ช่วยให้แยก concern ได้ชัด

### 5.1 POST /api/chat

**ใช้เป็น entrypoint เดียวสำหรับแชทบอท**

Request (ตัวอย่าง):

```json
{
  "message": "เดือนนี้ใช้ Shopee ไปเท่าไหร่",
  "image_path": null
}
```

หรือ

```json
{
  "message": "จ่ายค่าเคิมฉีหมาล่าเมื่อกี้นี้",
  "image_path": "C:/Users/nun/receipts/2026-01-30_2228.png"
}
```

Backend:
1. สร้าง context (ถ้ามี)
2. ส่งให้ LLM adapter → ได้ JSON: `intent`, filter, หรือ transaction fields
3. ถ้า `intent` เป็น:
   - `add_transaction` / `bill_payment` → validate + insert DB
   - `query_summary` → query DB แล้วประกอบคำตอบ
   - `unknown` → บอก user ว่าไม่เข้าใจ

Response (ตัวอย่าง):

```json
{
  "reply_text": "เดือนนี้ (2026-01) คุณใช้ Shopee ไป 1,280 บาท",
  "debug": {
    "intent": "query_summary",
    "filters": {
      "period": {
        "from": "2026-01-01",
        "to": "2026-01-31"
      },
      "channel": "shopee"
    }
  }
}
```

> หมายเหตุ: สำหรับ CLI-only MVP `image_path` อาจเป็น path บน local ที่ user ระบุเอง หรือ backend ไป map จาก ไฟล์ที่เลือกผ่าน UI ภายนอก

---

### 5.2 GET /api/dashboard/summary

ใช้ดึง metric ไปวาดกราฟ / หรือใช้ CLI ดูสรุปง่าย ๆ

```http
GET /api/dashboard/summary?from=2026-01-01&to=2026-01-31
```

Response example:

```json
{
  "period": {
    "from": "2026-01-01",
    "to": "2026-01-31"
  },
  "total_expense": 12340.0,
  "total_income": 15000.0,
  "by_category": [
    { "category": "food", "amount": 4200.0, "percent_of_expense": 34.0 },
    { "category": "shopping", "amount": 3100.0, "percent_of_expense": 25.1 },
    { "category": "bill", "amount": 2800.0, "percent_of_expense": 22.7 }
  ],
  "by_channel": [
    { "channel": "cash", "amount": 2000.0 },
    { "channel": "scb",  "amount": 5540.0 },
    { "channel": "tmw",  "amount": 2800.0 }
  ]
}
```

---

## 6. LLM Prompt Spec (Codex / อื่น ๆ)

### 6.1 Prompt สำหรับข้อความธรรมดา (add_transaction / query_summary)

**System Prompt (template):**

```text
You are a strict JSON parser for a single-user personal finance tracker.
User writes informal Thai messages about expenses and incomes.

You MUST respond with ONLY raw JSON. No explanation. No markdown.

Supported intents:
- "add_transaction": user logs a new income/expense/transfer.
- "query_summary": user asks for totals or breakdowns over some time period.
- "unknown": cannot confidently interpret the message.

When intent = "add_transaction", use this JSON format:

{
  "intent": "add_transaction",
  "transaction": {
    "txn_date": "YYYY-MM-DD or null",
    "amount": number or null,
    "currency": "THB",
    "direction": "income" | "expense" | "transfer",
    "channel": "cash" | "scb" | "kbank" | "tmw" | "unknown",
    "account_label": "string or null",
    "category": "food" | "rent" | "shopping" | "transport" | "bill" | "debt" | "other",
    "description": "string or null"
  }
}

When the user message is unclear or missing required information (like amount),
set those fields to null. NEVER guess.

When intent = "query_summary", use this JSON format:

{
  "intent": "query_summary",
  "filters": {
    "direction": "income" | "expense" | "both",
    "period": {
      "type": "month" | "day" | "range" | "year" | "all",
      "from": "YYYY-MM-DD or null",
      "to": "YYYY-MM-DD or null"
    },
    "category": "string or null",
    "channel": "string or null"
  }
}

If you really cannot understand, respond with:

{
  "intent": "unknown"
}
```

**User message example:**

```text
วันนี้กินข้าวมันไก่ไป 65 บาท เงินสด
```

Expected JSON (ตัวอย่าง):

```json
{
  "intent": "add_transaction",
  "transaction": {
    "txn_date": null,
    "amount": 65,
    "currency": "THB",
    "direction": "expense",
    "channel": "cash",
    "account_label": null,
    "category": "food",
    "description": "ข้าวมันไก่"
  }
}
```

---

### 6.2 Prompt สำหรับ OCR สลิป (bill_payment)

**System Prompt (template):**

```text
You are a strict JSON parser for Thai payment receipts (OCR text).
Dates are in Thai Buddhist calendar (พ.ศ.). Convert them to Gregorian (ค.ศ.) by subtracting 543.
Timezone is Asia/Bangkok.

You MUST respond with ONLY raw JSON. No explanation. No markdown.

Use this JSON format:

{
  "intent": "bill_payment",
  "transaction": {
    "txn_date": "YYYY-MM-DD or null",
    "amount": number or null,
    "currency": "THB",
    "direction": "expense",
    "channel": "tmw" | "scb" | "kbank" | "cash" | "unknown",
    "account_label": "string or null",
    "category": "food" | "bill" | "shopping" | "transport" | "other",
    "description": "string or null",
    "raw_ocr_text": "full original OCR text"
  }
}

Only fill fields when the information is clearly present or strongly implied.
If unclear, use null.
```

**User OCR text example (ย่อจากเคสจริงของคุณ):**

```text
จ่ายบิลสำเร็จ
30 ม.ค. 2569
22:28
รหัสอ้างอิง: 202601302kwhu jnbwv4zonhaz
จาก
นาย ฺชัชวาล อุปถัมภ์
xxx xxx749-9
โปยัง
truemoney
(เคิมฉีหม่าล่า
อุดมสุขวอร์ค)
biller Id
010554614953130
จำนวนเงิน
96.00
```

Expected JSON (ตัวอย่าง):

```json
{
  "intent": "bill_payment",
  "transaction": {
    "txn_date": "2026-01-30",
    "amount": 96.0,
    "currency": "THB",
    "direction": "expense",
    "channel": "tmw",
    "account_label": null,
    "category": "food",
    "description": "โปยัง (เคิมฉีหม่าล่า อุดมสุขวอร์ค) จ่ายบิลผ่าน TrueMoney",
    "raw_ocr_text": "จ่ายบิลสำเร็จ\n30 ม.ค. 2569\n22:28\nรหัสอ้างอิง: 202601302kwhu jnbwv4zonhaz\n..."
  }
}
```

---

## 7. Notes on Implementation (PC-only, minimal)

- ใช้ SQLite เป็นหลัก (ไฟล์เดียว พกพาง่าย)
- สคริปต์หลัก:
  - CLI/แชท: อ่าน stdin → เรียก `/api/chat` local → print `reply_text`
  - dashboard: static HTML + endpoint summary
- LLM adapter:
  - เรียก Codex ผ่าน CLI (`subprocess`) หรือ HTTP ก็ได้ แต่ต้อง:
    - บังคับ output เป็น JSON
    - ถ้า `json.loads()` ไม่ผ่าน → ถือว่า fail, ไม่เขียน DB
- ทุกครั้งที่ใช้ LLM:
  - เก็บ `raw_text` (สลิป / ข้อความ) ไว้เสมอ เพื่อ manual fix ทีหลัง
  - validation layer สำคัญกว่า prompt สวย
