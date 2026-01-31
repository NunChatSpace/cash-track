# personal-finance-tracker/spec.md

## 1. Goal

ลด friction ในการจดการเงินส่วนตัว  
โดยใช้ **รูปสลิปการโอนเงิน + โน้ตสั้น ๆ**  
แล้วให้ระบบช่วยแปลงเป็นข้อมูลที่บันทึกลงฐานข้อมูลได้

หลักคิด:
- ใช้แรงมนุษย์ให้น้อยที่สุด
- LLM ช่วย “ร่างข้อมูล” ไม่ใช่ “ตัดสินแทนมนุษย์”

---

## 2. Scope

### In-scope
- อัปโหลดรูปสลิป (bank / e-wallet / marketplace)
- แนบโน้ตสั้น ๆ (optional)
- ใช้ LLM (vision) ช่วย extract ข้อมูลเป็น JSON
- ให้ user ตรวจ + confirm ก่อนบันทึกลง DB

### Out-of-scope
- ไม่ auto-save โดยไม่ให้ user confirm
- ไม่ทำ budgeting / forecasting
- ไม่เชื่อม API ธนาคาร
- ไม่ต้องถูก 100%

---

## 3. High-level Flow

1. User อัปโหลดรูปสลิป
2. (optional) พิมพ์โน้ตสั้น ๆ
3. Backend:
   - เก็บรูป (S3 / MinIO / local)
   - ส่งรูป + note เข้า LLM พร้อม prompt
4. LLM ส่ง JSON กลับ
5. Backend validate JSON
6. Frontend แสดงฟอร์ม prefill
7. User แก้ / confirm
8. Save ลง DB (status = confirmed)

---

## 4. Data Model

```sql
CREATE TABLE transactions (
    id              BIGSERIAL PRIMARY KEY,
    txn_date        DATE            NOT NULL,
    amount          NUMERIC(12,2)   NOT NULL,
    currency        VARCHAR(10)     NOT NULL DEFAULT 'THB',
    direction       VARCHAR(10)     NOT NULL,
    channel         VARCHAR(50)     NOT NULL,
    "from"          TEXT,
    "to"            TEXT,
    category        VARCHAR(50),
    description     TEXT,
    slip_image_url  TEXT            NOT NULL,
    raw_ocr_text    TEXT,
    llm_confidence  NUMERIC(5,2),
    status          VARCHAR(20)     NOT NULL DEFAULT 'pending',
    created_at      TIMESTAMP       NOT NULL DEFAULT now(),
    updated_at      TIMESTAMP       NOT NULL DEFAULT now()
);
```

---

## 5. API Spec

### POST /api/transactions/slip
Upload slip + note เพื่อ parse ข้อมูล

### PATCH /api/transactions/:id/confirm
ยืนยันข้อมูลก่อนบันทึกจริง

---

## 6. LLM Prompt Spec

```text
You are a data-entry assistant for personal finance tracking.

Input:
1) An image of a payment slip (Thai bank / e-wallet / marketplace).
2) Optional user notes written informally.

Task:
- Extract ONLY information that is clearly visible or strongly implied.
- Do NOT guess.
- If unclear, return null.

Output JSON only with fields:
txn_date, amount, currency, direction, channel, from, to, category, description, confidence
```
