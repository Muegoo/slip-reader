# ocr-service — PaddleOCR หลัง HTTP สำหรับ slip-reader
#
# ไม่มีโค้ด Python ที่เขียนเองในนี้: ใช้ตัว serving ที่มากับ PaddleX แล้ว pin โมเดลผ่าน OCR.yaml
# เวอร์ชันตรงกับที่ใช้ตอน spike เพื่อให้ผลอ่านตรงกับ fixture ใน testdata/ocr
FROM python:3.10-slim

# libgomp = OpenMP runtime ที่ paddlepaddle ต้องการ / libgl+glib = ของ opencv ที่ paddleocr ใช้เปิดรูป
RUN apt-get update \
    && apt-get install -y --no-install-recommends libgomp1 libgl1 libglib2.0-0 \
    && rm -rf /var/lib/apt/lists/*

RUN pip install --no-cache-dir \
    paddlepaddle==3.3.1 \
    paddleocr==3.7.0 \
    "paddlex[serving]==3.7.2"

# จำกัด thread ตั้งแต่ตอนโหลดไลบรารี (spike: 4 thread พีค ~2.4 GB) และไม่ให้เช็คเน็ตหาโมเดลตอน start
ENV OMP_NUM_THREADS=4 \
    PADDLE_PDX_DISABLE_MODEL_SOURCE_CHECK=True

WORKDIR /app
COPY OCR.yaml /app/OCR.yaml

# โหลดโมเดลตอน build — container จะ start เร็วและทำงานได้โดยไม่ต้องต่อเน็ต
RUN python -c "from paddleocr import PaddleOCR; PaddleOCR(text_detection_model_name='PP-OCRv5_mobile_det', text_recognition_model_name='th_PP-OCRv5_mobile_rec', use_doc_orientation_classify=False, use_doc_unwarping=False, use_textline_orientation=False)"

EXPOSE 8866
CMD ["paddlex", "--serve", "--pipeline", "/app/OCR.yaml", "--host", "0.0.0.0", "--port", "8866"]
