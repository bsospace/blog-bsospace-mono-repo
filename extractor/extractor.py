from flask import Flask, request, jsonify
import fitz  # PyMuPDF
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
            full_text += page.get_text()
        return jsonify({"text": full_text})
    except Exception as e:
        return jsonify({"error": str(e)}), 500

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5002)
