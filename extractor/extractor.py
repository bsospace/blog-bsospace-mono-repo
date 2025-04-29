from flask import Flask, request, jsonify
import fitz  # PyMuPDF
import pytesseract
from pdf2image import convert_from_bytes
from io import BytesIO

app = Flask(__name__)

@app.route("/extract-text", methods=["POST"])
def extract_text():
    file = request.files.get('file')
    if not file:
        return jsonify({"error": "No file provided"}), 400

    try:
        data = file.read()
        pdf = fitz.open(stream=data, filetype="pdf")
        full_text = ""

        for page in pdf:
            text = page.get_text()
            if text.strip():  # ถ้าเจอ text
                full_text += text
            else:
                # ถ้าไม่เจอ text → OCR page นั้น
                images = convert_from_bytes(data, dpi=300, first_page=page.number+1, last_page=page.number+1)
                for image in images:
                    ocr_text = pytesseract.image_to_string(image, lang='tha+eng')
                    full_text += ocr_text

        return jsonify({"text": full_text})
    except Exception as e:
        return jsonify({"error": str(e)}), 500

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5002)
